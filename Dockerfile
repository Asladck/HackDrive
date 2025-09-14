FROM golang:1.25-alpine AS go-builder

WORKDIR /goapp
COPY ./cmd /goapp/cmd
COPY go.mod go.sum /goapp/
COPY ./cmd /goapp/cmd
COPY ./internal /goapp/internal
COPY ./scripts /goapp/scripts

RUN go mod init hackdrive || true
RUN go mod tidy
RUN go build -o /goapp/server ./cmd/api/main.go


FROM python:3.11-slim

RUN apt-get update && apt-get install -y \
    libglib2.0-0 \
    libsm6 \
    libxext6 \
    libxrender-dev \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY ./app.py /app/app.py
COPY ./requirements.txt /app/requirements.txt

RUN pip install --no-cache-dir torch==2.2.1+cpu torchvision==0.17.1+cpu --index-url https://download.pytorch.org/whl/cpu
RUN pip install --no-cache-dir --upgrade pip
RUN pip install --no-cache-dir -r /app/requirements.txt

COPY --from=go-builder /goapp/server /app/server

EXPOSE 5000 8000

CMD ["sh", "-c", "python /app/app.py & /app/server"]
