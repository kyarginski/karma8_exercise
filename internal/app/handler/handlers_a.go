package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"karma8/internal/app/processes"
	"karma8/internal/app/services"
	trccontext "karma8/internal/lib/context"
	"karma8/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GetFileItem(service services.IService) http.HandlerFunc {
	// swagger:operation GET /api/file/{id} GetFileItem
	// Get file from server by ID.
	// ---
	// description: Get file from server by ID.
	// parameters:
	// - name: id
	//   in: path
	//   description: The ID of the file.
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: OK
	//   '400':
	//     description: Bad User Request Error
	//   '404':
	//     description: File Not Found Error
	//   '500':
	//     description: Internal Server Error
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		ctx, span := trccontext.WithTelemetrySpan(r.Context(), "GetFileItem")
		defer span.End()

		// r = r.WithContext(ctx)

		span.SetTag("id", id)

		parsedUUID, err := uuid.Parse(id)
		if err != nil {
			http.Error(w, "Error parsing UUID", http.StatusBadRequest)

			return
		}

		data, err := service.GetFileItem(ctx, parsedUUID)
		if err != nil {
			http.Error(w, "error in GetFileItem: "+err.Error(), http.StatusInternalServerError)
			span.SetError(err)

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
	// swagger:operation PUT /api/file PutFileItem
	// Upload a file.
	// ---
	// description: Uploads a file to the server.
	// parameters:
	// - name: file
	//   in: formData
	//   description: The file to upload.
	//   required: true
	//   type: file
	// consumes:
	// - multipart/form-data
	// responses:
	//   '200':
	//     description: OK
	//     schema:
	//       "$ref": "#/definitions/ResponseSuccess"
	//   '400':
	//     description: Bad User Request Error
	//   '500':
	//     description: Internal Server Error
	return func(w http.ResponseWriter, r *http.Request) {
		// Парсинг формы с файлом
		err := r.ParseMultipartForm(10 << 20) // TODO (в настройки) 10 MB максимальный размер файла.
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		ctx, span := trccontext.WithTelemetrySpan(r.Context(), "PutFileItem")
		defer span.End()

		r = r.WithContext(ctx)

		// TODO debug
		span.AddEvent("установим event")
		span.SetTag("label1", "значение label1")

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
		// Запись файла на диск.
		_, err = processes.WriteFile(fileContent, handler.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			span.SetError(err)

			return
		}

		source := &models.FileItem{
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
