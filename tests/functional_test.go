package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"

	"karma8/internal/app"
	"karma8/internal/lib/logger/sl"
	"karma8/internal/models"
	"karma8/internal/testhelpers/postgres"
	"karma8/internal/testhelpers/redis"
	"karma8/internal/trace"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestHappyPath(t *testing.T) {
	// Подготовим тестовые данные.
	testFile := []byte(`ip_address,country_code,country,city,latitude,longitude,mystery_value
200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
160.103.7.140,CZ,Nicaragua,New Neva,-68.31023296602508,-37.62435199624531,7301823115
70.95.73.73,TL,Saudi Arabia,Gradymouth,-49.16675918861615,-86.05920084416894,2559997162
,PY,Falkland Islands (Malvinas),,75.41685191518815,-144.6943217219469,0
125.159.20.54,LI,Guyana,Port Karson,-78.2274228596799,-163.26218895343357,1337885276
not your IP address,HN,Benin,Fredyshire,-70.41275040993187,60.19866111663936,2040256925
`,
	)

	httpPort := 8260
	testDB, err := postgres.NewTestDatabase(t)
	if err != nil {
		fmt.Println("Error connecting to the database: ", err)
		t.Skip("This test is excluded from unit tests.")
	}
	assert.NotNil(t, testDB)
	t.Logf("Test conteiner postgres: %+v", testDB.DB().Stats())

	testRedis, err := redis.NewTestRedis(t)
	if err != nil {
		fmt.Println("Error connecting to the Redis: ", err)
		t.Skip("This test is excluded from unit tests.")
	}
	assert.NotNil(t, testRedis)
	t.Logf("Test container redis: %+v", testRedis.DB().ClientInfo(context.Background()).String())

	log := sl.SetupLogger("nop")

	tp, err := trace.InitJaegerTracer("http://localhost:14268/api/traces", "test-service_a", "nop", true)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Error(err.Error())
		}
	}(ctx)

	applicationA, err := app.NewServiceA(log, testDB.ConnectString(t), httpPort)
	defer applicationA.Stop()
	assert.NoError(t, err)

	go applicationA.Start()

	applicationB := make([]*app.App, 6)
	for i := 0; i < 6; i++ {
		applicationB[i], err = app.NewServiceB(log, testRedis.ConnectString(t), httpPort+1+i, i+1)
		defer applicationB[i].Stop()
		assert.NoError(t, err)

		go applicationB[i].Start()
	}

	// Сохраним файл на сервер.
	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)
	url := baseURL + "/api/file"

	// Создаем буфер для записи данных формы.
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Добавляем бинарные данные в теле формы.
	part, err := writer.CreateFormFile("file", "file")
	assert.NoError(t, err)
	_, err = part.Write(testFile)
	assert.NoError(t, err)

	// Закрываем тело формы
	err = writer.Close()
	assert.NoError(t, err)

	// Создаем HTTP запрос с методом PUT и устанавливаем заголовки
	request, err := http.NewRequest("PUT", url, &body)
	assert.NoError(t, err)

	request.Header.Set("Content-Type", writer.FormDataContentType())

	// Отправляем запрос
	response, err := http.DefaultClient.Do(request)
	assert.NoError(t, err)
	defer response.Body.Close()

	assert.Equal(t, http.StatusOK, response.StatusCode)

	got, err := io.ReadAll(response.Body)
	assert.NoError(t, err)

	var newID string
	if http.StatusOK == response.StatusCode {
		var gotResult models.ResponseSuccess

		err = json.Unmarshal(got, &gotResult)
		assert.NoError(t, err)
		assert.NotNil(t, gotResult)
		newID = gotResult.ID
		t.Logf("ID: %s", newID)
	} else {
		t.Errorf("Error: %s", got)
	}

	// Получим файл с сервера (из кеша).
	getFile(t, baseURL, newID, testFile)

	// Очистим файловый кеш.
	err = applicationA.ClearCacheAll()
	assert.NoError(t, err)

	// Получим файл с сервера (из Redis)
	getFile(t, baseURL, newID, testFile)
}

func getFile(t *testing.T, baseURL string, id string, testFile []byte) {
	t.Helper()

	// Получим файл с сервера (из кеша).
	url := fmt.Sprintf(baseURL+"/api/file/%s", id)

	response, err := http.Get(url)
	assert.NoError(t, err)
	defer response.Body.Close()

	// Проверяем код ответа.
	assert.Equal(t, http.StatusOK, response.StatusCode)

	// Читаем данные из тела ответа.
	data, err := io.ReadAll(response.Body)
	assert.NoError(t, err)

	diff := cmp.Diff(testFile, data)
	if diff != "" {
		t.Fatal("file result mismatch\n", diff)
	}
}
