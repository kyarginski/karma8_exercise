Golang Developer Test Cases
====

Вы решили создать конкурента Amazon S3 и знаете как сделать лучший сервис хранения файлов.

На сервер A по REST присылают файл, его надо разрезать на 6 примерно равных частей и сохранить на серверах хранения Bn (n ≥ 6).

При REST-запросе на сервер A нужно достать куски с серверов Bn склеить их и отдать файл.

# Имеем следующее
1. Один сервер для REST запросов
2. Несколько серверов для хранения кусков файлов

# Ограничения
1. Реализовать сервис и тестовый модуль, который обеспечит проверку его работы
2. Сервера для хранения могут добавляться в систему в любой момент, но не могут удаляться из системы
3. Предусмотреть равномерное заполнение серверов хранения
4. Код разместить в Github

# System schemes
![A](./doc/New_Amason_S3_service_A.svg)
![B](./doc/New_Amason_S3_service_B.svg)

# Схема базы данных

![Scheme DB](./doc/database_diagram.png)

# Соображения по архитектуре

## Реализация

Реализуем 2 приложения:
- service_a
- service_b

## Запуск service_a

```shell
export SERVICE_A_CONFIG_PATH=config/service_a/prod.yaml && go run ./cmd/service_a
```

## Запуск service_b

Нужно запустить несколько экземпляров service_b со своим портом, например:

```shell
export SERVICE_B_REDIS_DB=1 && export SERVICE_B_PORT=8261 && export SERVICE_B_CONFIG_PATH=config/service_b/local.yaml && go run ./cmd/service_b
```
```shell
export SERVICE_B_REDIS_DB=2 && export SERVICE_B_PORT=8262 && export SERVICE_B_CONFIG_PATH=config/service_b/local.yaml && go run ./cmd/service_b
```
```shell
export SERVICE_B_REDIS_DB=3 && export SERVICE_B_PORT=8263 && export SERVICE_B_CONFIG_PATH=config/service_b/local.yaml && go run ./cmd/service_b
```

## Запуск сервисов через docker containers

```shell
docker-compose up -d
```

## Останов сервисов в docker containers

```shell
docker-compose down
```

# Как тестировать с использованием Postman

### запрос на сохранение файла
```shell
PUT http://localhost:8260/api/file
```
тело запроса

- form-data

  - key: file
  - value: <выбранный файл>

Ответ

код 200 OK

тело ответа - id сохранённого файла
```json
{
    "id": "fe1f3f07-8eb3-11ee-829b-0242ac130006"
}
```

### запрос на получение файла
```shell
GET http://localhost:8260/api/file/{id}
```
где
id = fe1f3f07-8eb3-11ee-829b-0242ac130006

результат - содержимое файла.

Лучше всего тестировать на графических файлах (*.jpg, *.png etc.) так как на них хорошо видно соблюдение целостности файла при загрузке из частей.


# Добавляем новый сервер для хранения (bucket)

1) добавляем описание нового сервера в конфигурацию service_a
```sql
INSERT INTO bucket (id, address, active_sign) VALUES
    (7, 'http://host.docker.internal:8267', true);
``` 
2) запускаем новый экземпляр service_b
```shell
export SERVICE_B_REDIS_DB=7 && export SERVICE_B_PORT=8267 && export SERVICE_B_CONFIG_PATH=config/service_b/local.yaml && go run ./cmd/service_b
```

3) перезапускаем service_a

4) проверяем, что новый сервер участвует в сохранении файлов

```
-----------------
|   bucket_ids  |
-----------------
|{1,2,3,4,5,6,7}|
-----------------
```

## Подключение OpenTelemetry

https://www.jaegertracing.io/docs/1.47/getting-started/

Запуск
```
docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 14250:14250 \
  -p 14268:14268 \
  -p 14269:14269 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.51

```

Просмотр

http://localhost:16686 to access the Jaeger UI.


```
go get -u go.opentelemetry.io/otel
go get -u go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp

```

# Что ещё можно сделать

- более детальную обработку ошибок
- сделать удаление файлов с серверов хранения при удалении с сервера A
- сейчас кэш сделан в виде хранения файлов и очистка кэша сделана по таймеру; можно сделать другой механизм очистки кэша (например, по количеству файлов в кэше, LRU etc.) 
- вынести кэш в отдельный сервис (если хотим горизонтально масштабировать сервис A)
- сделать отдельный сервис управления бакетами (создание, удаление, список)
- ~~функциональные тесты с использованием testcontainer (postgres, redis)~~
- ~~добавить трассировку запросов через сервисы с использованием OpenTelemetry~~


Jaeger

https://www.inanzzz.com/index.php/post/4qes/implementing-opentelemetry-and-jaeger-tracing-in-golang-http-api
https://github.com/albertteoh/jaeger-go-example/blob/main/lib/tracing/init.go
