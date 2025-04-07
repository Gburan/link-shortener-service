# Link Shortener Service

## 1. Description

Сервис для сокращения URL-адресов

1. Метод `POST`, который сохраняет оригинальный URL в базе и возвращает сокращённый.
2. Метод `GET`, который принимает сокращённый URL и возвращает оригинальный URL.

## 2. Configuration

| Name           | Type    | Default value            | Description                    |
|----------------|---------|--------------------------|--------------------------------|
| SERVER_ADDRESS | String  | `:8080`                  | HTTP server address            |
| POSTGRES_CONN  | String  |                          | PostgreSQL connection string   |
| URL_LENGTH     | Integer | `10`                     | Length of generated short URLs |
| STORAGE_TYPE   | String  | `db`                     | Storage type (`db` or `map`)   |
| FIRST_URL_PART | String  | `https://somedomain.su/` | Base domain for short URLs     |

## 3. How to run
```
docker-compose up -d
```
или
```
make run-compose
```
или (с пересборкой контейнера)
```
make run-compose-b
```

## 4. Tests

### 4.1 Generate mocks
```
go generate ./...
```
или
```
make gen-mocks
```
### 4.2 Run tests
```
go test ./...
```
или 
```
make run-test-clean
```