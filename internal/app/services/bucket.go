package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	trccontext "karma8/internal/lib/context"
	"karma8/internal/lib/logger/sl"
	"karma8/internal/lib/middleware"
	"karma8/internal/models"

	"github.com/google/uuid"
)

const requestPath = "/api/filepart"

type Bucket struct {
	log    *slog.Logger
	client *http.Client
	path   string
	ID     int64
}

func NewBucket(log *slog.Logger, path string, id int64) *Bucket {
	return &Bucket{
		log: log,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		path: path + requestPath,
		ID:   id,
	}
}

// GetBucketsInfo возвращает информацию об активных бакетах.
func (s *Bucket) GetBucketsInfo() ([]*models.ServerBucketInfo, error) {
	result := make([]*models.ServerBucketInfo, 0)
	result = append(result, &models.ServerBucketInfo{
		ID:      s.ID,
		Address: s.path,
	})

	return result, nil
}

// SendToBucket отправляет файл в бакет.
func (s *Bucket) SendToBucket(ctx context.Context, item *models.BucketItem, id uuid.UUID) error {
	// Создаем буфер для записи данных формы.
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Добавляем поле ID.
	_ = writer.WriteField("id", id.String())

	// Добавляем бинарные данные в теле формы.
	part, err := writer.CreateFormFile("file", "file")
	if err != nil {
		return err
	}
	_, err = part.Write(item.Source)
	if err != nil {
		return err
	}
	// Закрываем тело формы
	_ = writer.Close()

	// Создаем HTTP запрос с методом PUT и устанавливаем заголовки
	request, err := http.NewRequestWithContext(ctx, "PUT", s.path, &body)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	// Отправляем запрос
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Обрабатываем ошибочный ответ.
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error in SendToBucket: %d", response.StatusCode)
	}

	return nil
}

// GetFromBucket получает части файла из бакета.
func (s *Bucket) GetFromBucket(ctx context.Context, id uuid.UUID, results map[int64][]byte, mutex *sync.Mutex) {
	url := fmt.Sprintf(s.path+"/%s", id)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.log.Error("GetFromBucket", "error creating request", err)
		return
	}

	requestID, ok := trccontext.RequestIDFromContext(ctx)
	if !ok {
		requestID = "UNKNOWN"
	}
	// Set the request-id header
	req.Header.Set(middleware.HeaderRequestID, requestID)

	// Perform the HTTP request
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		s.log.Error("GetFromBucket", sl.Err(err))
		return
	}
	defer response.Body.Close()

	// Check the response status code
	if response.StatusCode != http.StatusOK {
		s.log.Error("error status GetFromBucket",
			sl.Err(err),
			"status_code", response.StatusCode,
		)
		return
	}

	// Read data from the response body
	data, err := io.ReadAll(response.Body)
	if err != nil {
		s.log.Error("error io.ReadAll GetFromBucket",
			sl.Err(err),
		)
		return
	}

	// Save data to the map using the mutex
	mutex.Lock()
	results[s.ID] = data
	mutex.Unlock()
}
