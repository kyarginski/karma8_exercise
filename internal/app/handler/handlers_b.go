package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"karma8/internal/app/services"
	"karma8/internal/models"
	"karma8/internal/trace"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func GetBucketItem(service services.IService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		parsedUUID, err := uuid.Parse(id)
		if err != nil {
			http.Error(w, "Error parsing UUID", http.StatusBadRequest)
			return
		}

		var span oteltrace.Span
		if trace.UseTracing {
			span = trace.CreateSubSpan(context.Background(), r, trace.ServiceTitle)
			span.SetAttributes(
				attribute.String("func", "GetBucketItem"),
				attribute.String("id", id),
			)

			defer span.End()
		}

		data, err := service.GetFileItem(parsedUUID)
		if err != nil {
			http.Error(w, "error in GetFileItem", http.StatusInternalServerError)
			if span != nil {
				span.SetAttributes(
					attribute.String("error", err.Error()),
				)
			}

			return
		}

		// Устанавливаем заголовки.
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, data.FileName))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(len(data.FileContent)))

		// Отправляем содержимое файла.
		_, _ = w.Write(data.FileContent)
	}
}

func PutBucketItem(service services.IService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Парсинг формы с файлом
		err := r.ParseMultipartForm(10 << 20) // TODO (в настройки) 10 MB максимальный размер файла.
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		var span oteltrace.Span
		if trace.UseTracing {
			span = trace.CreateSubSpan(context.Background(), r, trace.ServiceTitle)
			span.SetAttributes(
				attribute.String("func", "GetFileItem"),
			)
			defer span.End()
		}

		// Получение файла из формы.
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to retrieve file from form", http.StatusBadRequest)
			if span != nil {
				span.SetAttributes(
					attribute.String("error", err.Error()),
				)
			}

			return
		}
		defer file.Close()

		// Получение содержимого файла из формы.
		fileContent, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file content", http.StatusInternalServerError)
			if span != nil {
				span.SetAttributes(
					attribute.String("error", err.Error()),
				)
			}

			return
		}
		// Получение ID файла из формы.
		fileID := r.FormValue("id")

		source := &models.FileItem{
			ID:              fileID,
			FileName:        handler.Filename,
			FileContentType: http.DetectContentType(fileContent),
			FileContent:     fileContent,
		}

		newID, err := service.PutFileItem(source)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			if span != nil {
				span.SetAttributes(
					attribute.String("error", err.Error()),
				)
			}

			return
		}

		w.Header().Set("Content-Type", "application/json")
		result := models.ResponseSuccess{
			ID: newID.String(),
		}
		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, "error in PutFileItem", http.StatusInternalServerError)
			if span != nil {
				span.SetAttributes(
					attribute.String("error", err.Error()),
				)
			}

			return
		}
	}
}
