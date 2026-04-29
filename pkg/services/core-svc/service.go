// Package coresvc implements the core API service that serves external and internal HTTP endpoints.
// It also consumes spr.package.updated messages to create collection and rebuild tasks,
// publishes spr.collection.requested and spr.rebuild.requested messages, and consumes
// spr.collection.completed and spr.rebuild.completed messages to update task status.
package coresvc

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"

	"git.duti.dev/secure-package-registry/internal/gen/coredb"
	"git.duti.dev/secure-package-registry/internal/messages"
	"git.duti.dev/secure-package-registry/pkg/logger"
	sprminio "git.duti.dev/secure-package-registry/pkg/minio"
	"git.duti.dev/secure-package-registry/pkg/pkgdb"
	"git.duti.dev/secure-package-registry/pkg/services"
	"git.duti.dev/secure-package-registry/pkg/services/core-svc/server"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

var log zerolog.Logger

func init() {
	log = logger.WithComponent("core-svc")
}

// Start runs the core-svc: HTTP servers, the spr.package.updated consumer,
// the spr.collection.completed consumer, and the spr.rebuild.completed consumer.
// Blocks until ctx is cancelled, then gracefully shuts down.
func Start(ctx context.Context, deps *services.Deps) error {
	db := pkgdb.NewClient(deps.Pool)
	queries := coredb.New(deps.Pool)

	publisher, err := amqp.NewPublisher(
		amqp.NewDurableQueueConfig(deps.Config.RabbitMQURL),
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return fmt.Errorf("creating AMQP publisher: %w", err)
	}
	defer func() {
		if cerr := publisher.Close(); cerr != nil {
			log.Warn().Err(cerr).Msg("Failed to close AMQP publisher")
		}
	}()

	// Subscriber for spr.package.updated.
	updatedSub, err := amqp.NewSubscriber(
		amqp.NewDurableQueueConfig(deps.Config.RabbitMQURL),
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return fmt.Errorf("creating package-updated subscriber: %w", err)
	}
	defer func() {
		if cerr := updatedSub.Close(); cerr != nil {
			log.Warn().Err(cerr).Msg("Failed to close package-updated subscriber")
		}
	}()

	updatedCh, err := updatedSub.Subscribe(ctx, "spr.package.updated")
	if err != nil {
		return fmt.Errorf("subscribing to package updates: %w", err)
	}

	// Subscriber for spr.collection.completed.
	completedSub, err := amqp.NewSubscriber(
		amqp.NewDurableQueueConfig(deps.Config.RabbitMQURL),
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return fmt.Errorf("creating collection-completed subscriber: %w", err)
	}
	defer func() {
		if cerr := completedSub.Close(); cerr != nil {
			log.Warn().Err(cerr).Msg("Failed to close collection-completed subscriber")
		}
	}()

	completedCh, err := completedSub.Subscribe(ctx, "spr.collection.completed")
	if err != nil {
		return fmt.Errorf("subscribing to collection completions: %w", err)
	}

	// Subscriber for spr.rebuild.completed.
	rebuildCompletedSub, err := amqp.NewSubscriber(
		amqp.NewDurableQueueConfig(deps.Config.RabbitMQURL),
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return fmt.Errorf("creating rebuild-completed subscriber: %w", err)
	}
	defer func() {
		if cerr := rebuildCompletedSub.Close(); cerr != nil {
			log.Warn().Err(cerr).Msg("Failed to close rebuild-completed subscriber")
		}
	}()

	rebuildCompletedCh, err := rebuildCompletedSub.Subscribe(ctx, "spr.rebuild.completed")
	if err != nil {
		return fmt.Errorf("subscribing to rebuild completions: %w", err)
	}

	// Run consumers in separate goroutines; report fatal errors via channels.
	updatedErr := make(chan error, 1)
	go func() {
		if err := consumePackageUpdated(ctx, queries, publisher, updatedCh); err != nil {
			updatedErr <- err
		}
	}()

	completedErr := make(chan error, 1)
	go func() {
		if err := consumeCollectionCompleted(ctx, queries, completedCh); err != nil {
			completedErr <- err
		}
	}()

	rebuildCompletedErr := make(chan error, 1)
	go func() {
		if err := consumeRebuildCompleted(ctx, queries, rebuildCompletedCh); err != nil {
			rebuildCompletedErr <- err
		}
	}()

	// Create MinIO client for admin artifact downloads.
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

	externalServer := server.NewExternal("0.0.0.0:"+deps.Config.CoreSvc.ExternalPort, db, server.AdminDeps{
		Querier:   queries,
		Publisher: publisher,
		MinIO:     minioClient,
	})
	internalServer := server.NewInternal("0.0.0.0:"+deps.Config.CoreSvc.InternalPort, queries, publisher)

	externalErrCh := externalServer.Start()
	internalErrCh := internalServer.Start()

	select {
	case <-ctx.Done():
		log.Info().Msg("Shutting down servers...")
	case err := <-externalErrCh:
		return fmt.Errorf("external server: %w", err)
	case err := <-internalErrCh:
		return fmt.Errorf("internal server: %w", err)
	case err := <-updatedErr:
		return fmt.Errorf("package-updated consumer: %w", err)
	case err := <-completedErr:
		return fmt.Errorf("collection-completed consumer: %w", err)
	case err := <-rebuildCompletedErr:
		return fmt.Errorf("rebuild-completed consumer: %w", err)
	}

	shutdownCtx := context.Background()
	var shutdownErr error
	if err := externalServer.Stop(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Failed to stop external server")
		shutdownErr = err
	}
	if err := internalServer.Stop(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Failed to stop internal server")
		if shutdownErr == nil {
			shutdownErr = err
		}
	}

	log.Info().Msg("Shutdown complete")
	return shutdownErr
}

// consumePackageUpdated reads spr.package.updated messages and, for each one:
//  1. Looks up the package by ecosystem+identifier.
//  2. Upserts the package version.
//  3. Creates and publishes a behavioral collection task.
//  4. Creates and publishes a rebuild verification task.
func consumePackageUpdated(
	ctx context.Context,
	queries *coredb.Queries,
	publisher message.Publisher,
	messagesCh <-chan *message.Message,
) error {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Stopping package-updated consumer")
			return nil
		case msg, ok := <-messagesCh:
			if !ok {
				log.Info().Msg("Package-updated subscription closed")
				return nil
			}
			if err := handlePackageUpdated(ctx, queries, publisher, msg); err != nil {
				log.Error().Err(err).Msg("Failed to handle package-updated message")
				msg.Nack()
				continue
			}
			msg.Ack()
		}
	}
}

func handlePackageUpdated(
	ctx context.Context,
	queries *coredb.Queries,
	publisher message.Publisher,
	msg *message.Message,
) error {
	var upd messages.PackageUpdated
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&upd); err != nil {
		return fmt.Errorf("decoding package-updated message: %w", err)
	}

	l := log.With().
		Str("ecosystem", upd.Ecosystem).
		Str("package", upd.Identifier).
		Str("version", upd.Version).
		Logger()

	// Look up the package. It must already exist because the poller created it.
	pkg, err := queries.GetPackageByEcosystemAndIdentifier(ctx, coredb.GetPackageByEcosystemAndIdentifierParams{
		Ecosystem:  coredb.Ecosystem(upd.Ecosystem),
		Identifier: upd.Identifier,
	})
	if err != nil {
		return fmt.Errorf("looking up package %s/%s: %w", upd.Ecosystem, upd.Identifier, err)
	}

	// Upsert the package version. source_url is null here; the COALESCE in the
	// query preserves any existing value.
	pvID, err := queries.InsertPackageVersion(ctx, coredb.InsertPackageVersionParams{
		PackageID: pkg.ID,
		Version:   upd.Version,
		SourceUrl: pgtype.Text{}, // null
	})
	if err != nil {
		return fmt.Errorf("upserting package version: %w", err)
	}

	// Update the package's latest_version to reflect the newly detected version.
	err = queries.UpdatePackageLatestVersion(ctx, coredb.UpdatePackageLatestVersionParams{
		ID:            pkg.ID,
		LatestVersion: pgtype.Text{String: upd.Version, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("updating package latest version: %w", err)
	}

	// Dedup: skip if a collection task is already pending or running.
	active, err := queries.HasActiveCollectionTask(ctx, coredb.HasActiveCollectionTaskParams{
		PackageVersionID: pvID,
		Source:           "npm",
	})
	if err != nil {
		return fmt.Errorf("checking active collection task: %w", err)
	}
	if active {
		l.Debug().Msg("Active collection task already exists, skipping")
	} else {
		// Insert the collection task. ON CONFLICT DO NOTHING handles races.
		task, err := queries.InsertCollectionTask(ctx, coredb.InsertCollectionTaskParams{
			PackageVersionID: pvID,
			Source:           "npm",
		})
		if errors.Is(err, pgx.ErrNoRows) {
			// Conflict — another consumer beat us to it.
			l.Debug().Msg("Collection task already exists (conflict), skipping")
		} else if err != nil {
			return fmt.Errorf("inserting collection task: %w", err)
		} else {
			// Publish spr.collection.requested for be-runner to pick up.
			collReq := messages.CollectionRequested{
				TaskID:     task.ID,
				Ecosystem:  upd.Ecosystem,
				Identifier: upd.Identifier,
				Version:    upd.Version,
			}
			var buf bytes.Buffer
			if err := gob.NewEncoder(&buf).Encode(collReq); err != nil {
				return fmt.Errorf("encoding collection-requested message: %w", err)
			}
			wmMsg := message.NewMessage(watermill.NewUUID(), buf.Bytes())
			if err := publisher.Publish("spr.collection.requested", wmMsg); err != nil {
				// Log but don't fail — the task is persisted; a retry mechanism can
				// re-publish later.
				l.Error().Err(err).Msg("Failed to publish collection-requested (task persisted)")
			} else {
				l.Info().Int32("task_id", task.ID).Msg("Created collection task and published collection request")
			}
		}
	}

	if err := createAndPublishRebuildTask(ctx, queries, publisher, pvID, upd); err != nil {
		l.Error().Err(err).Msg("Failed to create or publish rebuild task")
	}

	l.Info().Msg("Handled package update")
	return nil
}

func createAndPublishRebuildTask(
	ctx context.Context,
	queries *coredb.Queries,
	publisher message.Publisher,
	pvID int32,
	upd messages.PackageUpdated,
) error {
	l := log.With().
		Str("ecosystem", upd.Ecosystem).
		Str("package", upd.Identifier).
		Str("version", upd.Version).
		Str("source", "oss-rebuild").
		Logger()

	active, err := queries.HasActiveRebuildTask(ctx, coredb.HasActiveRebuildTaskParams{
		PackageVersionID: pvID,
		Source:           "oss-rebuild",
	})
	if err != nil {
		return fmt.Errorf("checking active rebuild task: %w", err)
	}
	if active {
		l.Debug().Msg("Active rebuild task already exists, skipping")
		return nil
	}

	task, err := queries.InsertRebuildTask(ctx, coredb.InsertRebuildTaskParams{
		PackageVersionID: pvID,
		Source:           "oss-rebuild",
	})
	if errors.Is(err, pgx.ErrNoRows) {
		l.Debug().Msg("Rebuild task already exists (conflict), skipping")
		return nil
	}
	if err != nil {
		return fmt.Errorf("inserting rebuild task: %w", err)
	}

	req := messages.RebuildRequested{
		TaskID:     task.ID,
		Ecosystem:  upd.Ecosystem,
		Identifier: upd.Identifier,
		Version:    upd.Version,
		Source:     "oss-rebuild",
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(req); err != nil {
		return fmt.Errorf("encoding rebuild-requested message: %w", err)
	}

	wmMsg := message.NewMessage(watermill.NewUUID(), buf.Bytes())
	if err := publisher.Publish("spr.rebuild.requested", wmMsg); err != nil {
		l.Error().Err(err).Msg("Failed to publish rebuild-requested (task persisted)")
		return nil
	}

	l.Info().Int32("task_id", task.ID).Msg("Created rebuild task and published rebuild request")
	return nil
}

// consumeCollectionCompleted reads spr.collection.completed messages from be-runner
// and updates collection task status accordingly.
func consumeCollectionCompleted(
	ctx context.Context,
	queries *coredb.Queries,
	messagesCh <-chan *message.Message,
) error {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Stopping collection-completed consumer")
			return nil
		case msg, ok := <-messagesCh:
			if !ok {
				log.Info().Msg("Collection-completed subscription closed")
				return nil
			}
			if err := handleCollectionCompleted(ctx, queries, msg); err != nil {
				log.Error().Err(err).Msg("Failed to handle collection-completed message")
				msg.Nack()
				continue
			}
			msg.Ack()
		}
	}
}

func handleCollectionCompleted(
	ctx context.Context,
	queries *coredb.Queries,
	msg *message.Message,
) error {
	var completed messages.CollectionCompleted
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&completed); err != nil {
		return fmt.Errorf("decoding collection-completed message: %w", err)
	}

	l := log.With().
		Int32("task_id", completed.TaskID).
		Str("ecosystem", completed.Ecosystem).
		Str("package", completed.Identifier).
		Str("version", completed.Version).
		Logger()

	if completed.Success {
		err := queries.UpdateCollectionTaskSucceeded(ctx, coredb.UpdateCollectionTaskSucceededParams{
			ID:             completed.TaskID,
			ArtifactBucket: pgtype.Text{String: completed.ArtifactBucket, Valid: true},
			ArtifactKey:    pgtype.Text{String: completed.ArtifactKey, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("marking task %d succeeded: %w", completed.TaskID, err)
		}
		l.Info().
			Str("bucket", completed.ArtifactBucket).
			Str("key", completed.ArtifactKey).
			Msg("Collection task succeeded")
	} else {
		err := queries.UpdateCollectionTaskFailed(ctx, coredb.UpdateCollectionTaskFailedParams{
			ID:            completed.TaskID,
			FailureReason: pgtype.Text{String: completed.FailureReason, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("marking task %d failed: %w", completed.TaskID, err)
		}
		l.Warn().
			Str("reason", completed.FailureReason).
			Msg("Collection task failed")
	}

	return nil
}

// consumeRebuildCompleted reads spr.rebuild.completed messages from rebuild-worker
// and updates rebuild task status accordingly.
func consumeRebuildCompleted(
	ctx context.Context,
	queries *coredb.Queries,
	messagesCh <-chan *message.Message,
) error {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Stopping rebuild-completed consumer")
			return nil
		case msg, ok := <-messagesCh:
			if !ok {
				log.Info().Msg("Rebuild-completed subscription closed")
				return nil
			}
			if err := handleRebuildCompleted(ctx, queries, msg); err != nil {
				log.Error().Err(err).Msg("Failed to handle rebuild-completed message")
				msg.Nack()
				continue
			}
			msg.Ack()
		}
	}
}

func handleRebuildCompleted(
	ctx context.Context,
	queries *coredb.Queries,
	msg *message.Message,
) error {
	var completed messages.RebuildCompleted
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&completed); err != nil {
		return fmt.Errorf("decoding rebuild-completed message: %w", err)
	}

	l := log.With().
		Int32("task_id", completed.TaskID).
		Str("ecosystem", completed.Ecosystem).
		Str("package", completed.Identifier).
		Str("version", completed.Version).
		Str("source", completed.Source).
		Logger()

	if completed.Success {
		err := queries.UpdateRebuildTaskSucceeded(ctx, coredb.UpdateRebuildTaskSucceededParams{
			ID:                     completed.TaskID,
			Matched:                pgtype.Bool{Bool: completed.Matched, Valid: true},
			OfficialArtifactBucket: textOrNull(completed.OfficialArtifactBucket),
			OfficialArtifactKey:    textOrNull(completed.OfficialArtifactKey),
			RebuiltArtifactBucket:  textOrNull(completed.RebuiltArtifactBucket),
			RebuiltArtifactKey:     textOrNull(completed.RebuiltArtifactKey),
			DiffoscopeBucket:       textOrNull(completed.DiffoscopeBucket),
			DiffoscopeKey:          textOrNull(completed.DiffoscopeKey),
			LogsBucket:             textOrNull(completed.LogsBucket),
			LogsKey:                textOrNull(completed.LogsKey),
			MetadataBucket:         textOrNull(completed.MetadataBucket),
			MetadataKey:            textOrNull(completed.MetadataKey),
		})
		if err != nil {
			return fmt.Errorf("marking rebuild task %d succeeded: %w", completed.TaskID, err)
		}
		l.Info().
			Bool("matched", completed.Matched).
			Msg("Rebuild task succeeded")
		return nil
	}

	if completed.Unavailable {
		err := queries.UpdateRebuildTaskUnavailable(ctx, coredb.UpdateRebuildTaskUnavailableParams{
			ID:            completed.TaskID,
			FailureReason: textOrNull(completed.FailureReason),
		})
		if err != nil {
			return fmt.Errorf("marking rebuild task %d unavailable: %w", completed.TaskID, err)
		}
		l.Warn().
			Str("reason", completed.FailureReason).
			Msg("Rebuild task unavailable")
		return nil
	}

	err := queries.UpdateRebuildTaskFailed(ctx, coredb.UpdateRebuildTaskFailedParams{
		ID:            completed.TaskID,
		FailureReason: textOrNull(completed.FailureReason),
	})
	if err != nil {
		return fmt.Errorf("marking rebuild task %d failed: %w", completed.TaskID, err)
	}

	l.Warn().
		Str("reason", completed.FailureReason).
		Msg("Rebuild task failed")

	return nil
}

func textOrNull(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}
