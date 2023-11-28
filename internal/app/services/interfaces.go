package services

import (
	"karma8/internal/models"

	"github.com/google/uuid"
)

type IService interface {
	GetFileItem(id uuid.UUID) (*models.FileItem, error)
	PutFileItem(source *models.FileItem) (uuid.UUID, error)
	DeleteFileItem(id uuid.UUID) error
	Close() error
}

type IBucket interface {
	GetFilePart(id uuid.UUID) ([]byte, error)
	PutFilePart(id uuid.UUID, source []byte) error
	DeleteFilePart(id uuid.UUID) error
}
