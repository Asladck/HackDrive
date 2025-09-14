# -------- Stage 1: Go build --------
FROM golang:1.25-alpine AS go-builder

WORKDIR /goapp
COPY ./cmd /goapp/cmd
RUN go mod init hackdrive || true
RUN go mod tidy
RUN go build -o /goapp/server ./cmd/api/main.go

# -------- Stage 2: Python --------
FROM python:3.11-slim

# Устанавливаем зависимости ОС
RUN apt-get update && apt-get install -y \
    build-essential \
    libgl1 \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Копируем Python файлы
COPY ./app.py /app/app.py
COPY ./requirements.txt /app/requirements.txt

# Устанавливаем зависимости Python
RUN pip install --no-cache-dir --upgrade pip
RUN pip install --no-cache-dir -r /app/requirements.txt

# Копируем Go-сервер из первого этапа
COPY --from=go-builder /goapp/server /app/server

# Экспонируем порты
EXPOSE 5000 8000  # 5000 для FastAPI, 8000 для Go (если нужно)

# Скрипт запуска обоих сервисов
CMD ["sh", "-c", "python /app/app.py & /app/server"]
