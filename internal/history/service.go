package history

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/phillezi/server-room-temperature/internal/dto"
)

const (
	MaxHistory     = 31 * 24 * time.Hour
	DefaultHistory = time.Hour
	DefaultSubject = "serverroom.temperature.room1.sensor1"

	// At 1 Hz:
	//   1 hour  ~= 3,600 messages
	//   1 day   ~= 86,400 messages
	//   31 days ~= 2.68M messages
	//
	// Large batches significantly reduce request/response overhead.
	FetchBatchSize = 30_000
)

type Service struct {
	stream jetstream.Stream
}

func New(stream jetstream.Stream) *Service {
	return &Service{
		stream: stream,
	}
}

func (s *Service) History(
	ctx context.Context,
	subject string,
	from time.Time,
	to time.Time,
) ([]dto.Reading, error) {
	if subject == "" {
		subject = DefaultSubject
	}

	from = from.UTC()
	to = to.UTC()

	if !to.After(from) {
		return nil, fmt.Errorf("from must be before to")
	}

	if to.Sub(from) > MaxHistory {
		return nil, fmt.Errorf(
			"history range exceeds maximum of %s",
			MaxHistory,
		)
	}

	// Your sensor currently produces one reading per second.
	// This gives us a useful initial allocation without wasting
	// enormous amounts of memory for normal requests.
	capacity := int(to.Sub(from).Seconds()) + 1

	readings := make([]dto.Reading, 0, capacity)

	consumer, err := s.stream.CreateConsumer(
		ctx,
		jetstream.ConsumerConfig{
			Description:   "HTTP history query",
			FilterSubject: subject,
			DeliverPolicy: jetstream.DeliverByStartTimePolicy,
			OptStartTime:  &from,
			AckPolicy:     jetstream.AckNonePolicy,
			ReplayPolicy:  jetstream.ReplayInstantPolicy,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create history consumer: %w",
			err,
		)
	}

	defer func() {
		deleteCtx, cancel := context.WithTimeout(
			context.Background(),
			time.Second,
		)
		defer cancel()

		_ = s.stream.DeleteConsumer(deleteCtx, consumer.CachedInfo().Name)
	}()

	for {
		batch, err := consumer.FetchNoWait(FetchBatchSize)
		if err != nil {
			if errors.Is(err, context.Canceled) ||
				errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}

			return nil, fmt.Errorf(
				"fetch history: %w",
				err,
			)
		}

		count := 0

		for msg := range batch.Messages() {
			count++

			var reading dto.Reading

			if err := json.Unmarshal(
				msg.Data(),
				&reading,
			); err != nil {
				// Ignore malformed historical messages.
				continue
			}

			// The consumer starts at `from`, but this protects us
			// against any timestamp mismatch in the payload.
			if reading.Timestamp.Before(from) {
				continue
			}

			if reading.Timestamp.After(to) {
				continue
			}

			readings = append(readings, reading)
		}

		if err := batch.Error(); err != nil {
			return nil, fmt.Errorf(
				"read history batch: %w",
				err,
			)
		}

		// FetchNoWait returns only what is currently available.
		// Once nothing is returned, the historical backlog is drained.
		if count == 0 {
			break
		}

		// Don't spin unnecessarily when the final batch was smaller
		// than requested. There cannot be more than this batch if
		// FetchNoWait drained everything currently available.
		if count < FetchBatchSize {
			break
		}
	}

	return readings, nil
}
