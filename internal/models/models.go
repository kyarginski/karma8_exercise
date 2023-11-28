package models

import (
	"time"

	"github.com/google/uuid"
)

// FileItem - структура для хранения записи файла.
type FileItem struct {
	ID              string `json:"id"`
	FileName        string `json:"file_name"`
	FileContentType string `json:"file_content_type"`
	FileContent     []byte `json:"file_content"`
}

// BucketItem - структура для хранения элемента корзины.
type BucketItem struct {
	ID     int64  `json:"id"`
	Source []byte `json:"source"`
}

// ServerBucketInfo - структура для хранения информации о сервере корзины.
type ServerBucketInfo struct {
	ID      int64  `json:"id"`
	Address string `json:"string"`
}

// CacheItem - структура для работы с таблицей cache в базе данных.
type CacheItem struct {
	Checksum  string    `json:"checksum" db:"checksum"`
	FileName  string    `json:"filename" db:"filename"`
	ExpiredAt time.Time `json:"expired_at" db:"expired_at"`
}

// MetadataItem - структура для таблицы metadata.
type MetadataItem struct {
	UUID        uuid.UUID `db:"uuid" json:"uuid"`
	Checksum    string    `db:"checksum" json:"checksum"`
	FileName    string    `db:"filename" json:"filename"`
	ContentType string    `db:"content_type" json:"content_type"`
	BucketIDs   []int64   `db:"bucket_ids" json:"bucket_ids"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// ResponseSuccess - структура для возврата ответа об успешном сохранении файла.
type ResponseSuccess struct {
	ID string `json:"id"`
}

// ResponseError - структура для возврата ответа об ошибке.
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
