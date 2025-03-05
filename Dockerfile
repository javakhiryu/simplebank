# Используем образ Go версии 1.24.0 на базе Alpine Linux (легковесный дистрибутив) для этапа сборки.
# Этап называется "builder", чтобы его можно было использовать в следующем этапе.
FROM golang:1.24.0-alpine3.21 AS builder

# Устанавливаем рабочую директорию внутри контейнера как `/app`.
WORKDIR /app

# Копируем все файлы из текущей директории (на хосте) в рабочую директорию контейнера (`/app`).
COPY . .

# Компилируем Go-приложение, создавая бинарный файл `main`.
RUN go build -o main main.go

# Устанавливаем утилиту `curl` для загрузки файлов.
RUN apk add curl

# Повторная установка `curl` без кэширования (избыточно, так как уже установлено выше).
#RUN apk --no-cache add curl

# Скачиваем и распаковываем утилиту `migrate` для управления миграциями базы данных.
#RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.2/migrate.linux-amd64.tar.gz | tar xvz

# Делаем скрипт `start.sh` исполняемым.
RUN chmod +x start.sh

# Этап запуска (run)
# Используем легковесный образ Alpine Linux для финального этапа.
FROM alpine:3.21

# Устанавливаем рабочую директорию внутри контейнера как `/app`.
WORKDIR /app

# Копируем скомпилированный бинарный файл `main` из этапа `builder`.
COPY --from=builder /app/main .

# Копируем утилиту `migrate` из этапа `builder`.
#COPY --from=builder /app/migrate ./migrate

# Копируем файл с переменными окружения (`app.env`) в контейнер.
COPY app.env .

# Копируем скрипт `start.sh` в контейнер.
COPY start.sh .

# Копируем скрипт `wait-for.sh` в контейнер.
COPY wait-for.sh .

# Копируем папку с миграциями базы данных в контейнер.
COPY db/migration ./db/migration

# Открываем порт 8080 для доступа к приложению.
EXPOSE 8080

# Указываем команду по умолчанию для запуска приложения.
CMD ["/app/main"]

# Указываем точку входа (entrypoint) для запуска скрипта `start.sh`.
ENTRYPOINT [ "/app/start.sh" ]