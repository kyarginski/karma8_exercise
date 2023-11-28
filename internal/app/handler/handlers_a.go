package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"karma8/internal/app/processes"
	"karma8/internal/app/services"
	"karma8/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GetFileItem(service services.IService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fileID := vars["id"]
		parsedUUID, err := uuid.Parse(fileID)
		if err != nil {
			http.Error(w, "Error parsing UUID", http.StatusBadRequest)
			return
		}

		data, err := service.GetFileItem(parsedUUID)
		if err != nil {
			http.Error(w, "error in GetFileItem", http.StatusInternalServerError)
			return
		}

		// Устанавливаем заголовки.
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, data.FileName))
		w.Header().Set("Content-Type", data.FileContentType)
		w.Header().Set("Content-Length", strconv.Itoa(len(data.FileContent)))

		// Отправляем содержимое файла.
		_, _ = w.Write(data.FileContent)
	}
}

func PutFileItem(service services.IService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Парсинг формы с файлом
		err := r.ParseMultipartForm(10 << 20) // TODO (в настройки) 10 MB максимальный размер файла.
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		// Получение файла из формы.
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to retrieve file from form", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Получение содержимого файла из формы.
		fileContent, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file content", http.StatusInternalServerError)
			return
		}
		// Запись файла на диск.
		_, err = processes.WriteFile(fileContent, handler.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		source := &models.FileItem{
			FileName:        handler.Filename,
			FileContentType: http.DetectContentType(fileContent),
			FileContent:     fileContent,
		}

		newID, err := service.PutFileItem(source)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		result := models.ResponseSuccess{
			ID: newID.String(),
		}
		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, "error in PutFileItem", http.StatusInternalServerError)
			return
		}
	}
}
