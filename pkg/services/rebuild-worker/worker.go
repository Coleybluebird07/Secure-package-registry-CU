// Package rebuildworker implements the reproducible-build verification worker.
//
// This worker is intentionally separate from the behavioral-analysis runner.
// It consumes spr.rebuild.requested messages and publishes spr.rebuild.completed
// messages. The actual OSS Rebuild and diffoscope implementation is added in
// the next feature issue; this issue establishes the queue and worker boundary.
package rebuildworker

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"git.duti.dev/secure-package-registry/internal/messages"
	"git.duti.dev/secure-package-registry/pkg/logger"
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

			log.Info().
				Int32("task_id", req.TaskID).
				Str("ecosystem", req.Ecosystem).
				Str("package", req.Identifier).
				Str("version", req.Version).
				Str("source", req.Source).
				Msg("Received rebuild verification request")

			completed := messages.RebuildCompleted{
				TaskID:        req.TaskID,
				Ecosystem:     req.Ecosystem,
				Identifier:    req.Identifier,
				Version:       req.Version,
				Source:        req.Source,
				Unavailable:   true,
				FailureReason: "rebuild verification worker skeleton only; OSS Rebuild integration not implemented yet",
			}

			publishCompletion(publisher, completed)
			msg.Ack()
		}
	}
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
