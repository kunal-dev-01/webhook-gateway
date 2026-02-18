package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"shared/logger"
)

const avatarQueueKey = "queue:avatar:upload"

// ============================================================================
// Job
// ============================================================================

type AvatarUploadJob struct {
	UserID      uuid.UUID `json:"user_id"`
	ObjectKey   string    `json:"object_key"`
	FileData    []byte    `json:"file_data"`
	ContentType string    `json:"content_type"`
	EnqueuedAt  time.Time `json:"enqueued_at"`
}

// ============================================================================
// Queue
// ============================================================================

type AvatarUploadQueue struct {
	redis *redis.Client
}

func NewAvatarUploadQueue(redisClient *redis.Client) *AvatarUploadQueue {
	return &AvatarUploadQueue{redis: redisClient}
}

// Enqueue pushes an avatar upload job to Redis list
func (q *AvatarUploadQueue) Enqueue(ctx context.Context, job AvatarUploadJob) error {
	job.EnqueuedAt = time.Now()

	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal avatar job: %w", err)
	}

	return q.redis.LPush(ctx, avatarQueueKey, data).Err()
}

// Dequeue pops a job from queue (blocking, 5s timeout)
func (q *AvatarUploadQueue) Dequeue(ctx context.Context) (*AvatarUploadJob, error) {
	result, err := q.redis.BRPop(ctx, 5*time.Second, avatarQueueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // timeout, no job
		}
		return nil, fmt.Errorf("failed to dequeue avatar job: %w", err)
	}

	var job AvatarUploadJob
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal avatar job: %w", err)
	}

	return &job, nil
}

// Len returns current queue length
func (q *AvatarUploadQueue) Len(ctx context.Context) (int64, error) {
	return q.redis.LLen(ctx, avatarQueueKey).Result()
}

// ============================================================================
// Worker
// ============================================================================

type AvatarWorkerDeps interface {
	UploadToR2(ctx context.Context, key string, data []byte, contentType string) (string, error)
	SaveAvatarURL(ctx context.Context, userID uuid.UUID, url string) error
}

// StartAvatarWorker starts a background goroutine that processes avatar upload jobs.
// Call this once from main.go. Pass a cancel context to stop the worker gracefully.
func StartAvatarWorker(ctx context.Context, queue *AvatarUploadQueue, r2 *R2Client, saveURL func(ctx context.Context, userID uuid.UUID, url string) error) {
	go func() {
		logger.Info("🖼️  Avatar upload worker started")

		for {
			select {
			case <-ctx.Done():
				logger.Info("🛑 Avatar upload worker stopped")
				return
			default:
				job, err := queue.Dequeue(ctx)
				if err != nil {
					logger.Error("Avatar worker: dequeue error", err)
					continue
				}
				if job == nil {
					continue // timeout, loop again
				}

				processAvatarJob(ctx, job, r2, saveURL)
			}
		}
	}()
}

func processAvatarJob(ctx context.Context, job *AvatarUploadJob, r2 *R2Client, saveURL func(ctx context.Context, userID uuid.UUID, url string) error) {
	uploadCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url, err := r2.Upload(uploadCtx, job.ObjectKey, bytes.NewReader(job.FileData), job.ContentType)
	if err != nil {
		logger.Error("Avatar worker: R2 upload failed", err, "user_id", job.UserID, "key", job.ObjectKey)
		return
	}

	if err := saveURL(uploadCtx, job.UserID, url); err != nil {
		logger.Error("Avatar worker: DB save failed", err, "user_id", job.UserID, "url", url)
		return
	}

	logger.Info("Avatar uploaded successfully", "user_id", job.UserID, "url", url)
}