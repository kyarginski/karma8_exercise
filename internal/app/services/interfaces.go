package services

import (
	"context"
	"log/slog"

	"karma8/internal/models"

	"github.com/google/uuid"
)

type IService interface {
	Logger() *slog.Logger
	Ping(ctx context.Context) bool
	GetFileItem(ctx context.Context, id uuid.UUID) (*models.FileItem, error)
	PutFileItem(ctx context.Context, source *models.FileItem) (uuid.UUID, error)
	DeleteFileItem(ctx context.Context, id uuid.UUID) error
	ClearCacheAll() error
	Close() error
}
