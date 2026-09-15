package rebuildworker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"git.duti.dev/secure-package-registry/internal/messages"
)

func TestReadFileWithLimitAllowsFileWithinLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "artifact.txt")

	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	data, err := readFileWithLimit(path, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected hello, got %q", string(data))
	}
}

func TestReadFileWithLimitRejectsOversizedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "artifact.txt")

	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := readFileWithLimit(path, 4)
	if err == nil {
		t.Fatal("expected size limit error")
	}
	if !strings.Contains(err.Error(), "exceeds limit") {
		t.Fatalf("expected size limit error, got %v", err)
	}
}

func TestBuildDiffoscopeReportUsesRealReportEvenWhenDiffoscopeExitsNonZero(t *testing.T) {
	dir := t.TempDir()

	diffoscopeScript := filepath.Join(dir, "fake-diffoscope")
	outputPath := filepath.Join(dir, "diffoscope.html")
	officialPath := filepath.Join(dir, "official.tgz")
	rebuiltPath := filepath.Join(dir, "rebuilt.tgz")

	if err := os.WriteFile(officialPath, []byte("official"), 0o644); err != nil {
		t.Fatalf("write official: %v", err)
	}
	if err := os.WriteFile(rebuiltPath, []byte("rebuilt"), 0o644); err != nil {
		t.Fatalf("write rebuilt: %v", err)
	}

	script := `#!/bin/sh
set -eu
# args: --html output official rebuilt
cat > "$2" <<'HTML'
<!doctype html>
<html><body><h1>real diffoscope report</h1></body></html>
HTML
exit 1
`
	if err := os.WriteFile(diffoscopeScript, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	report, fallback := buildDiffoscopeReport(
		context.Background(),
		diffoscopeScript,
		5*time.Second,
		1024*1024,
		messages.RebuildRequested{
			TaskID:     1,
			Ecosystem:  "npm",
			Identifier: "chalk",
			Version:    "5.6.2",
			Source:     "oss-rebuild",
		},
		officialPath,
		rebuiltPath,
		outputPath,
		"officialhash",
		"rebuilthash",
	)

	if fallback {
		t.Fatal("expected real report, got fallback")
	}
	if !strings.Contains(string(report), "real diffoscope report") {
		t.Fatalf("expected real report content, got %s", string(report))
	}
}

func TestBuildDiffoscopeReportFallsBackWhenNoReportIsCreated(t *testing.T) {
	dir := t.TempDir()

	diffoscopeScript := filepath.Join(dir, "fake-diffoscope")
	outputPath := filepath.Join(dir, "diffoscope.html")
	officialPath := filepath.Join(dir, "official.tgz")
	rebuiltPath := filepath.Join(dir, "rebuilt.tgz")

	if err := os.WriteFile(officialPath, []byte("official"), 0o644); err != nil {
		t.Fatalf("write official: %v", err)
	}
	if err := os.WriteFile(rebuiltPath, []byte("rebuilt"), 0o644); err != nil {
		t.Fatalf("write rebuilt: %v", err)
	}

	script := `#!/bin/sh
echo "diffoscope failed badly" >&2
exit 2
`
	if err := os.WriteFile(diffoscopeScript, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	report, fallback := buildDiffoscopeReport(
		context.Background(),
		diffoscopeScript,
		5*time.Second,
		1024*1024,
		messages.RebuildRequested{
			TaskID:     2,
			Ecosystem:  "npm",
			Identifier: "chalk",
			Version:    "5.6.2",
			Source:     "oss-rebuild",
		},
		officialPath,
		rebuiltPath,
		outputPath,
		"officialhash",
		"rebuilthash",
	)

	if !fallback {
		t.Fatal("expected fallback report")
	}
	if !strings.Contains(string(report), "Diffoscope fallback report") {
		t.Fatalf("expected fallback report, got %s", string(report))
	}
	if !strings.Contains(string(report), "diffoscope failed badly") {
		t.Fatalf("expected captured stderr in fallback report, got %s", string(report))
	}
}

func TestBuildDiffoscopeReportFallsBackOnTimeout(t *testing.T) {
	dir := t.TempDir()

	diffoscopeScript := filepath.Join(dir, "fake-diffoscope")
	outputPath := filepath.Join(dir, "diffoscope.html")
	officialPath := filepath.Join(dir, "official.tgz")
	rebuiltPath := filepath.Join(dir, "rebuilt.tgz")

	if err := os.WriteFile(officialPath, []byte("official"), 0o644); err != nil {
		t.Fatalf("write official: %v", err)
	}
	if err := os.WriteFile(rebuiltPath, []byte("rebuilt"), 0o644); err != nil {
		t.Fatalf("write rebuilt: %v", err)
	}

	script := `#!/bin/sh
sleep 1
`
	if err := os.WriteFile(diffoscopeScript, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	report, fallback := buildDiffoscopeReport(
		context.Background(),
		diffoscopeScript,
		1*time.Nanosecond,
		1024*1024,
		messages.RebuildRequested{
			TaskID:     3,
			Ecosystem:  "npm",
			Identifier: "chalk",
			Version:    "5.6.2",
			Source:     "oss-rebuild",
		},
		officialPath,
		rebuiltPath,
		outputPath,
		"officialhash",
		"rebuilthash",
	)

	if !fallback {
		t.Fatal("expected fallback report")
	}
	if !strings.Contains(string(report), "timed out") {
		t.Fatalf("expected timeout reason in fallback report, got %s", string(report))
	}
}

func TestRunOSSRebuildTimeout(t *testing.T) {
	dir := t.TempDir()
	officialPath := filepath.Join(dir, "official")
	rebuiltPath := filepath.Join(dir, "rebuilt")

	if err := os.WriteFile(officialPath, []byte("official"), 0o644); err != nil {
		t.Fatalf("write official: %v", err)
	}

	logText, err := runOSSRebuild(
		context.Background(),
		"sleep 1",
		1*time.Nanosecond,
		messages.RebuildRequested{
			TaskID:     4,
			Ecosystem:  "npm",
			Identifier: "chalk",
			Version:    "5.6.2",
			Source:     "oss-rebuild",
		},
		"https://example.invalid/artifact.tgz",
		officialPath,
		rebuiltPath,
	)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if logText != "" {
		t.Fatalf("expected empty log text, got %q", logText)
	}
}

func TestFallbackDiffoscopeHTMLIncludesUsefulDetails(t *testing.T) {
	report := fallbackDiffoscopeHTML(
		messages.RebuildRequested{
			TaskID:     99,
			Ecosystem:  "pypi",
			Identifier: "requests",
			Version:    "2.33.1",
			Source:     "oss-rebuild",
		},
		"diffoscope",
		[]string{"--html", "out.html", "official", "rebuilt"},
		"officialhash",
		"rebuilthash",
		[]byte("captured output"),
		errors.New("diffoscope failed"),
	)

	text := string(report)

	for _, expected := range []string{
		"Diffoscope fallback report",
		"pypi",
		"requests",
		"2.33.1",
		"officialhash",
		"rebuilthash",
		"diffoscope --html out.html official rebuilt",
		"diffoscope failed",
		"captured output",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected fallback report to contain %q, got %s", expected, text)
		}
	}
}

func TestDownloadOfficialArtifactUnsupportedGo(t *testing.T) {
	_, _, err := downloadOfficialArtifact(
		context.Background(),
		"https://registry.npmjs.org",
		"go",
		"github.com/pkg/errors",
		"v0.9.1",
	)

	if err == nil {
		t.Fatal("expected unsupported Go error")
	}
	if !errors.Is(err, errRebuildUnavailable) {
		t.Fatalf("expected errRebuildUnavailable, got %v", err)
	}
	if !strings.Contains(err.Error(), "Go modules") {
		t.Fatalf("expected Go modules message, got %v", err)
	}
}

func TestDownloadOfficialArtifactUnsupportedEcosystem(t *testing.T) {
	_, _, err := downloadOfficialArtifact(
		context.Background(),
		"https://registry.npmjs.org",
		"rubygems",
		"rails",
		"1.0.0",
	)

	if err == nil {
		t.Fatal("expected unsupported ecosystem error")
	}
	if !errors.Is(err, errRebuildUnavailable) {
		t.Fatalf("expected errRebuildUnavailable, got %v", err)
	}
	if !strings.Contains(err.Error(), "unsupported ecosystem") {
		t.Fatalf("expected unsupported ecosystem message, got %v", err)
	}
}
