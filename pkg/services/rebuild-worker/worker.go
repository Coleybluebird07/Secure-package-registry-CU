// Package rebuildworker implements the reproducible-build verification worker.
//
// The worker consumes spr.rebuild.requested messages, downloads the official
// npm artifact, attempts to obtain a rebuilt artifact through a configurable
// OSS Rebuild command hook, compares both artifacts, runs diffoscope when they
// differ, stores outputs in MinIO, and publishes spr.rebuild.completed.
package rebuildworker

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"git.duti.dev/secure-package-registry/internal/messages"
	"git.duti.dev/secure-package-registry/pkg/logger"
	sprminio "git.duti.dev/secure-package-registry/pkg/minio"
	"git.duti.dev/secure-package-registry/pkg/services"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	watermillmsg "github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
)

var log zerolog.Logger

func init() {
	log = logger.WithComponent("rebuild-worker")
}

type npmMetadata struct {
	Name     string                `json:"name"`
	Versions map[string]npmVersion `json:"versions"`
}

type npmVersion struct {
	Version string  `json:"version"`
	Dist    npmDist `json:"dist"`
}

type npmDist struct {
	Tarball   string `json:"tarball"`
	Shasum    string `json:"shasum,omitempty"`
	Integrity string `json:"integrity,omitempty"`
}

type rebuildMetadata struct {
	TaskID          int32     `json:"task_id"`
	Ecosystem       string    `json:"ecosystem"`
	Package         string    `json:"package"`
	Version         string    `json:"version"`
	Source          string    `json:"source"`
	Matched         bool      `json:"matched"`
	OfficialSHA256  string    `json:"official_sha256"`
	RebuiltSHA256   string    `json:"rebuilt_sha256,omitempty"`
	OfficialTarball string    `json:"official_tarball"`
	GeneratedAt     time.Time `json:"generated_at"`
}

// Start runs the rebuild verification worker.
// It blocks until ctx is cancelled.
func Start(ctx context.Context, deps *services.Deps) error {
	subscriber, err := amqp.NewSubscriber(
		amqp.NewDurableQueueConfig(deps.Config.RabbitMQURL),
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return fmt.Errorf("creating AMQP subscriber: %w", err)
	}
	defer func() {
		if closeErr := subscriber.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("Failed to close AMQP subscriber")
		}
	}()

	publisher, err := amqp.NewPublisher(
		amqp.NewDurableQueueConfig(deps.Config.RabbitMQURL),
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return fmt.Errorf("creating AMQP publisher: %w", err)
	}
	defer func() {
		if closeErr := publisher.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("Failed to close AMQP publisher")
		}
	}()

	minioCfg := deps.Config.MinIO
	minioClient, err := sprminio.NewClient(ctx, sprminio.Config{
		Endpoint:  minioCfg.Endpoint,
		AccessKey: minioCfg.AccessKey,
		SecretKey: minioCfg.SecretKey,
		UseSSL:    minioCfg.UseSSL,
		Bucket:    minioCfg.Bucket,
	})
	if err != nil {
		return fmt.Errorf("creating minio client: %w", err)
	}

	messagesCh, err := subscriber.Subscribe(ctx, "spr.rebuild.requested")
	if err != nil {
		return fmt.Errorf("subscribing to rebuild requests: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Shutting down rebuild-worker")
			return nil
		case msg, ok := <-messagesCh:
			if !ok {
				log.Info().Msg("Rebuild subscription closed")
				return nil
			}

			var req messages.RebuildRequested
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&req); err != nil {
				log.Error().Err(err).Msg("Failed to decode rebuild request")
				msg.Nack()
				continue
			}

			completed := handleRebuildRequest(ctx, deps, minioClient, req)
			publishCompletion(publisher, completed)
			msg.Ack()
		}
	}
}

func handleRebuildRequest(
	ctx context.Context,
	deps *services.Deps,
	minioClient *sprminio.Client,
	req messages.RebuildRequested,
) messages.RebuildCompleted {
	l := log.With().
		Int32("task_id", req.TaskID).
		Str("ecosystem", req.Ecosystem).
		Str("package", req.Identifier).
		Str("version", req.Version).
		Str("source", req.Source).
		Logger()

	l.Info().Msg("Received rebuild verification request")

	completed := messages.RebuildCompleted{
		TaskID:     req.TaskID,
		Ecosystem:  req.Ecosystem,
		Identifier: req.Identifier,
		Version:    req.Version,
		Source:     req.Source,
	}

	if req.Ecosystem != "npm" {
		completed.Unavailable = true
		completed.FailureReason = fmt.Sprintf("rebuild verification only supports npm in this worker, got %q", req.Ecosystem)
		return completed
	}

	workDir, cleanup, err := createWorkDir(deps.Config.Rebuild.WorkDir, req)
	if err != nil {
		completed.FailureReason = err.Error()
		return completed
	}
	defer cleanup()

	officialPath := filepath.Join(workDir, "official.tgz")
	rebuiltPath := filepath.Join(workDir, "rebuilt.tgz")
	diffPath := filepath.Join(workDir, "diffoscope.html")
	logPath := filepath.Join(workDir, "rebuild.log")

	tarballURL, officialBytes, err := downloadOfficialNPMTarball(ctx, deps.Config.NPM.RegistryURL, req.Identifier, req.Version)
	if err != nil {
		completed.FailureReason = fmt.Sprintf("downloading official npm tarball: %v", err)
		return completed
	}
	if err := os.WriteFile(officialPath, officialBytes, 0o644); err != nil {
		completed.FailureReason = fmt.Sprintf("writing official tarball: %v", err)
		return completed
	}

	officialHash := sha256Hex(officialBytes)

	if deps.Config.Rebuild.OSSRebuildCmd == "" {
		completed.Unavailable = true
		completed.FailureReason = "OSS_REBUILD_CMD is not configured"
		_ = storeUnavailableMetadata(ctx, minioClient, req, tarballURL, officialHash, completed.FailureReason)
		return completed
	}

	rebuildLog, err := runOSSRebuild(ctx, deps.Config.Rebuild.OSSRebuildCmd, req, tarballURL, officialPath, rebuiltPath)
	_ = os.WriteFile(logPath, []byte(rebuildLog), 0o644)
	if err != nil {
		if errors.Is(err, errRebuildUnavailable) {
			completed.Unavailable = true
		}
		completed.FailureReason = err.Error()
		storeLogs(ctx, minioClient, req, logPath, &completed)
		return completed
	}

	rebuiltBytes, err := os.ReadFile(rebuiltPath)
	if err != nil {
		completed.FailureReason = fmt.Sprintf("reading rebuilt artifact: %v", err)
		storeLogs(ctx, minioClient, req, logPath, &completed)
		return completed
	}

	rebuiltHash := sha256Hex(rebuiltBytes)
	matched := bytes.Equal(officialBytes, rebuiltBytes)

	baseKey := rebuildBaseKey(req)
	officialKey := baseKey + "/official.tgz"
	rebuiltKey := baseKey + "/rebuilt.tgz"
	metadataKey := baseKey + "/metadata.json"
	logsKey := baseKey + "/rebuild.log"

	if err := minioClient.PutObject(ctx, officialKey, officialBytes, "application/gzip"); err != nil {
		completed.FailureReason = fmt.Sprintf("uploading official artifact: %v", err)
		return completed
	}
	if err := minioClient.PutObject(ctx, rebuiltKey, rebuiltBytes, "application/gzip"); err != nil {
		completed.FailureReason = fmt.Sprintf("uploading rebuilt artifact: %v", err)
		return completed
	}

	metadata := rebuildMetadata{
		TaskID:          req.TaskID,
		Ecosystem:       req.Ecosystem,
		Package:         req.Identifier,
		Version:         req.Version,
		Source:          req.Source,
		Matched:         matched,
		OfficialSHA256:  officialHash,
		RebuiltSHA256:   rebuiltHash,
		OfficialTarball: tarballURL,
		GeneratedAt:     time.Now().UTC(),
	}
	metadataBytes, _ := json.MarshalIndent(metadata, "", "  ")
	if err := minioClient.PutObject(ctx, metadataKey, metadataBytes, "application/json"); err != nil {
		completed.FailureReason = fmt.Sprintf("uploading rebuild metadata: %v", err)
		return completed
	}

	if _, err := os.Stat(logPath); err == nil {
		if data, readErr := os.ReadFile(logPath); readErr == nil {
			_ = minioClient.PutObject(ctx, logsKey, data, "text/plain")
			completed.LogsBucket = minioClient.Bucket()
			completed.LogsKey = logsKey
		}
	}

	completed.Success = true
	completed.Matched = matched
	completed.OfficialArtifactBucket = minioClient.Bucket()
	completed.OfficialArtifactKey = officialKey
	completed.RebuiltArtifactBucket = minioClient.Bucket()
	completed.RebuiltArtifactKey = rebuiltKey
	completed.MetadataBucket = minioClient.Bucket()
	completed.MetadataKey = metadataKey

	if !matched {
		diffBytes, err := runDiffoscope(ctx, deps.Config.Rebuild.DiffoscopeCmd, officialPath, rebuiltPath, diffPath)
		if err != nil {
			l.Warn().Err(err).Msg("Diffoscope failed; rebuild mismatch will still be recorded")
			diffBytes = []byte(fmt.Sprintf("diffoscope failed: %v\n", err))
		}
		diffKey := baseKey + "/diffoscope.html"
		contentType := "text/html"
		if !strings.Contains(string(diffBytes[:min(len(diffBytes), 128)]), "<") {
			contentType = "text/plain"
		}
		if err := minioClient.PutObject(ctx, diffKey, diffBytes, contentType); err != nil {
			completed.FailureReason = fmt.Sprintf("uploading diffoscope report: %v", err)
			completed.Success = false
			return completed
		}
		completed.DiffoscopeBucket = minioClient.Bucket()
		completed.DiffoscopeKey = diffKey
	}

	l.Info().
		Bool("matched", matched).
		Str("official_sha256", officialHash).
		Str("rebuilt_sha256", rebuiltHash).
		Msg("Rebuild verification completed")

	return completed
}

func createWorkDir(root string, req messages.RebuildRequested) (string, func(), error) {
	if root == "" {
		root = os.TempDir()
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", func() {}, fmt.Errorf("creating rebuild workdir root: %w", err)
	}

	dir, err := os.MkdirTemp(root, fmt.Sprintf("spr-rebuild-%d-", req.TaskID))
	if err != nil {
		return "", func() {}, fmt.Errorf("creating rebuild workdir: %w", err)
	}

	return dir, func() { _ = os.RemoveAll(dir) }, nil
}

func downloadOfficialNPMTarball(ctx context.Context, registryURL, packageName, version string) (string, []byte, error) {
	metadataURL, err := npmMetadataURL(registryURL, packageName)
	if err != nil {
		return "", nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL, nil)
	if err != nil {
		return "", nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("fetching npm metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("fetching npm metadata returned status %d", resp.StatusCode)
	}

	var meta npmMetadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return "", nil, fmt.Errorf("decoding npm metadata: %w", err)
	}

	versionData, ok := meta.Versions[version]
	if !ok {
		return "", nil, fmt.Errorf("version %s not found in npm metadata for %s", version, packageName)
	}
	if versionData.Dist.Tarball == "" {
		return "", nil, fmt.Errorf("npm metadata has no tarball URL for %s@%s", packageName, version)
	}

	tarballReq, err := http.NewRequestWithContext(ctx, http.MethodGet, versionData.Dist.Tarball, nil)
	if err != nil {
		return "", nil, err
	}
	tarballResp, err := http.DefaultClient.Do(tarballReq)
	if err != nil {
		return "", nil, fmt.Errorf("downloading npm tarball: %w", err)
	}
	defer tarballResp.Body.Close()

	if tarballResp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("downloading npm tarball returned status %d", tarballResp.StatusCode)
	}

	data, err := io.ReadAll(tarballResp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("reading npm tarball: %w", err)
	}

	return versionData.Dist.Tarball, data, nil
}

func npmMetadataURL(registryURL, packageName string) (string, error) {
	base, err := url.Parse(strings.TrimRight(registryURL, "/") + "/")
	if err != nil {
		return "", fmt.Errorf("invalid npm registry URL: %w", err)
	}

	escaped := packageName
	if strings.HasPrefix(packageName, "@") {
		escaped = strings.Replace(packageName, "/", "%2F", 1)
	} else {
		escaped = url.PathEscape(packageName)
	}

	rel, err := url.Parse(escaped)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(rel).String(), nil
}

var errRebuildUnavailable = errors.New("oss rebuild unavailable")

func runOSSRebuild(
	ctx context.Context,
	command string,
	req messages.RebuildRequested,
	tarballURL string,
	officialPath string,
	rebuiltPath string,
) (string, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Env = append(os.Environ(),
		"SPR_PACKAGE="+req.Identifier,
		"SPR_VERSION="+req.Version,
		"SPR_ECOSYSTEM="+req.Ecosystem,
		"SPR_SOURCE="+req.Source,
		"SPR_OFFICIAL_TARBALL_URL="+tarballURL,
		"SPR_OFFICIAL_ARTIFACT="+officialPath,
		"SPR_REBUILT_ARTIFACT="+rebuiltPath,
	)

	output, err := cmd.CombinedOutput()
	logText := string(output)

	if err != nil {
		return logText, fmt.Errorf("%w: command failed: %v: %s", errRebuildUnavailable, err, strings.TrimSpace(logText))
	}

	if _, err := os.Stat(rebuiltPath); err != nil {
		return logText, fmt.Errorf("%w: command did not create SPR_REBUILT_ARTIFACT at %s", errRebuildUnavailable, rebuiltPath)
	}

	return logText, nil
}

func runDiffoscope(ctx context.Context, diffoscopeCmd, officialPath, rebuiltPath, outputPath string) ([]byte, error) {
	if diffoscopeCmd == "" {
		diffoscopeCmd = "diffoscope"
	}

	cmd := exec.CommandContext(ctx, diffoscopeCmd, "--html", outputPath, officialPath, rebuiltPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("running diffoscope: %w: %s", err, strings.TrimSpace(string(output)))
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		return output, fmt.Errorf("reading diffoscope output: %w", err)
	}

	return data, nil
}

func storeLogs(ctx context.Context, minioClient *sprminio.Client, req messages.RebuildRequested, logPath string, completed *messages.RebuildCompleted) {
	data, err := os.ReadFile(logPath)
	if err != nil || len(data) == 0 {
		return
	}
	key := rebuildBaseKey(req) + "/rebuild.log"
	if err := minioClient.PutObject(ctx, key, data, "text/plain"); err != nil {
		log.Warn().Err(err).Msg("Failed to store rebuild log")
		return
	}
	completed.LogsBucket = minioClient.Bucket()
	completed.LogsKey = key
}

func storeUnavailableMetadata(ctx context.Context, minioClient *sprminio.Client, req messages.RebuildRequested, tarballURL, officialHash, reason string) error {
	metadata := map[string]any{
		"task_id":          req.TaskID,
		"ecosystem":        req.Ecosystem,
		"package":          req.Identifier,
		"version":          req.Version,
		"source":           req.Source,
		"official_tarball": tarballURL,
		"official_sha256":  officialHash,
		"unavailable":      true,
		"reason":           reason,
		"generated_at":     time.Now().UTC(),
	}
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return minioClient.PutObject(ctx, rebuildBaseKey(req)+"/metadata.json", data, "application/json")
}

func rebuildBaseKey(req messages.RebuildRequested) string {
	return fmt.Sprintf(
		"rebuild/%s/%s/%s/%s",
		safePath(req.Ecosystem),
		safePath(req.Identifier),
		safePath(req.Version),
		safePath(req.Source),
	)
}

func safePath(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "/", "__")
	value = strings.ReplaceAll(value, "@", "")
	value = strings.ReplaceAll(value, " ", "_")
	if value == "" {
		return "unknown"
	}
	return value
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func publishCompletion(publisher watermillmsg.Publisher, completed messages.RebuildCompleted) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(completed); err != nil {
		log.Error().Err(err).Int32("task_id", completed.TaskID).Msg("Failed to encode rebuild completion event")
		return
	}

	msg := watermillmsg.NewMessage(watermill.NewUUID(), buf.Bytes())
	if err := publisher.Publish("spr.rebuild.completed", msg); err != nil {
		log.Error().Err(err).Int32("task_id", completed.TaskID).Msg("Failed to publish rebuild completion event")
	}
}
