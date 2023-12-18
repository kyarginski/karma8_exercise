package services

import (
	"context"
	"log/slog"
	"time"

	"karma8/internal/app/health"
	"karma8/internal/models"

	"github.com/google/uuid"
)

type IService interface {
	Logger() *slog.Logger
	Ping(ctx context.Context) bool
	GetFileItem(ctx context.Context, id uuid.UUID) (*models.FileItem, error)
	PutFileItem(ctx context.Context, source *models.FileItem) (uuid.UUID, error)
	DeleteFileItem(ctx context.Context, id uuid.UUID) error
	ClearCache(d time.Duration)
	ClearCacheAll() error
	Close() error

	health.LivenessChecker
	health.ReadinessChecker
}
