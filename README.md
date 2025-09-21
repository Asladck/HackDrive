# inHack ML Service (Go + Python + gRPC + FastAPI/YOLO)

## 📌 Описание проекта

Этот проект состоит из двух частей:

1. **Python ML-сервис**  
   - Использует [YOLO](https://github.com/ultralytics/ultralytics) для детекции объектов.  
   - Обёрнут в gRPC-сервис, который принимает изображение Машин и возвращает предсказания царапин ржавчины и чистоты(bounding boxes, confidence, классы).  

2. **Go Backend**  
   - HTTP-сервер (через `router` и `handler`).  
   - gRPC-клиент для вызова Python ML-сервиса.  
   - Может использовать Python предсказания внутри своих API-эндпоинтов.  

Таким образом, Go остаётся основным backend, а Python выполняет только ML-задачи.

---

## ⚙️ Архитектура

Client (Postman / Frontend)
|
v
[ Go HTTP API ] ----> [ Python gRPC Server (YOLO model) ]

├── mlservice_pb2_grpc.py
└── my_model.pt # обученная YOLO модель
## 🔧 Установка и генерация

### 1. Установить зависимости для Go
```bash
go mod tidy
```
Установить protoc плагины:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
2. Установить зависимости для Python
```bash
protoc --go_out=. --go-grpc_out=. proto/mlservice.proto
```
3. Сгенерировать код из .proto
```bash
protoc --go_out=. --go-grpc_out=. proto/mlservice.proto
```
Файлы появятся в internal/mlservicepb/.

# ▶️ Запуск
1. Запустить Python ML-сервис
```bash
cd python_service
python server.py
```

Сервис поднимется на localhost:50051.

2. Запустить Go Backend
```bash
go run main.go
```

HTTP-сервер запустится на порту из configs/config.yaml (например, :8080).

✅ Функционал

 Загрузка изображения через gRPC

 YOLO предсказания (bounding boxes, confidence, classes)

 Go вызывает Python gRPC и может отдавать результат по HTTP

 Добавить Docker для Go + Python сервисов

 Настроить CI/CD
