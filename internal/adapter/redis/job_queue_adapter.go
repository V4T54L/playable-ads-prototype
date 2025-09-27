package redis

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const renderQueueKey = "render_queue"

type JobQueueAdapter struct {
	client *redis.Client
}

func NewJobQueueAdapter(client *redis.Client) *JobQueueAdapter {
	return &JobQueueAdapter{client: client}
}

func (q *JobQueueAdapter) Enqueue(ctx context.Context, jobID uuid.UUID) error {
	return q.client.LPush(ctx, renderQueueKey, jobID.String()).Err()
}

func (q *JobQueueAdapter) Dequeue(ctx context.Context) (uuid.UUID, error) {
	// Use BRPOP for a blocking pop. Timeout is 0, so it blocks indefinitely.
	result, err := q.client.BRPop(ctx, 0, renderQueueKey).Result()
	if err != nil {
		// redis.Nil is returned on timeout, but with 0 it shouldn't happen unless context is cancelled.
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, nil // Or a specific error indicating timeout/empty
		}
		return uuid.Nil, err
	}

	// result is a slice: [key, value]
	if len(result) != 2 {
		return uuid.Nil, errors.New("invalid response from redis BRPop")
	}

	jobID, err := uuid.Parse(result[1])
	if err != nil {
		return uuid.Nil, err
	}

	return jobID, nil
}
