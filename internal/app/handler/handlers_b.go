package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"karma8/internal/app/services"
	libcontext "karma8/internal/lib/context"
	"karma8/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GetBucketItem(service services.IService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		ctx, span := libcontext.WithTelemetrySpan(r.Context(), "GetBucketItem")
		defer span.End()

		// r = r.WithContext(ctx)

		span.SetTag("id", id)

		parsedUUID, err := uuid.Parse(id)
		if err != nil {
			http.Error(w, "Error parsing UUID", http.StatusBadRequest)
			span.SetError(err)

			return
		}

		data, err := service.GetFileItem(ctx, parsedUUID)
		if err != nil {
			http.Error(w, "error in GetFileItem", http.StatusInternalServerError)
			span.SetError(err)

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
		ctx, span := libcontext.WithTelemetrySpan(r.Context(), "PutBucketItem")
		defer span.End()

		r = r.WithContext(ctx)

		// Парсинг формы с файлом
		err := r.ParseMultipartForm(10 << 20) // TODO (в настройки) 10 MB максимальный размер файла.
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			span.SetError(err)

			return
		}

		// Получение файла из формы.
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to retrieve file from form", http.StatusBadRequest)
			span.SetError(err)

			return
		}
		defer file.Close()

		// Получение содержимого файла из формы.
		fileContent, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file content", http.StatusInternalServerError)
			span.SetError(err)

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

		newID, err := service.PutFileItem(ctx, source)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			span.SetError(err)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		result := models.ResponseSuccess{
			ID: newID.String(),
		}
		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, "error in PutFileItem", http.StatusInternalServerError)
			span.SetError(err)

			return
		}
	}
}
