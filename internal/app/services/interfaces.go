package services

import (
	"karma8/internal/models"

	"github.com/google/uuid"
)

type IService interface {
	GetFileItem(id uuid.UUID) (*models.FileItem, error)
	PutFileItem(source *models.FileItem) (uuid.UUID, error)
	DeleteFileItem(id uuid.UUID) error
	ClearCacheAll() error
	Close() error
}
