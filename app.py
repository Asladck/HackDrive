from fastapi import FastAPI, File, UploadFile
from fastapi.middleware.cors import CORSMiddleware
from ultralytics import YOLO
import cv2
import numpy as np
import uvicorn
from typing import List, Dict, Any
import os

app = FastAPI(title="HackDrive ML Service")

# Настройка CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Загрузка модели
model = YOLO("yolov8n.pt")  # Можно заменить на свою модель

@app.get("/")
async def root():
    return {"message": "HackDrive ML Service is running!"}

@app.get("/health")
async def health_check():
    return {"status": "healthy"}

@app.post("/predict")
async def predict(file: UploadFile = File(...)):
    try:
        # Читаем файл
        contents = await file.read()
        np_img = np.frombuffer(contents, np.uint8)
        img = cv2.imdecode(np_img, cv2.IMREAD_COLOR)

        if img is None:
            return {"error": "Invalid image data", "predictions": []}

        # Делаем предсказание
        results = model(img)

        # Обрабатываем результаты
        predictions = []
        for r in results:
            if r.boxes is not None:
                # Конвертируем тензоры в списки
                boxes = r.boxes.xyxy.cpu().numpy().tolist()
                scores = r.boxes.conf.cpu().numpy().tolist()
                classes = r.boxes.cls.cpu().numpy().astype(int).tolist()

                prediction_data = {
                    "boxes": boxes,
                    "scores": scores,
                    "classes": classes
                }
                predictions.append(prediction_data)
            else:
                predictions.append({
                    "boxes": [],
                    "scores": [],
                    "classes": []
                })

        return {
            "predictions": predictions,
            "status": "success",
            "objects_detected": sum(len(pred["boxes"]) for pred in predictions)
        }

    except Exception as e:
        return {"error": str(e), "predictions": []}

if __name__ == "__main__":
    port = int(os.getenv("PORT", 5000))
    uvicorn.run(app, host="0.0.0.0", port=port, log_level="info")