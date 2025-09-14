import os
import time
import json
import random
from pathlib import Path
from typing import List, Optional

import cv2
import numpy as np
import torch
from torchvision.ops import nms
from ultralytics import YOLO
from fastapi import FastAPI, File, UploadFile, Form
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

# ========== Конфиг ==========
MODEL_PATHS = [
    "model1/best.pt",
    "model2/best.pt",
    "model3/best.pt",
]
RESULTS_DIR = "results"
os.makedirs(RESULTS_DIR, exist_ok=True)

# ========== FastAPI ==========
app = FastAPI(title="YOLO Triple Ensemble Service")
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# ========== Ансамбль ==========
class TripleEnsemble:
    def __init__(self, model_paths: List[str]):
        if len(model_paths) != 3:
            raise ValueError("Нужно ровно 3 модели")
        self.model_paths = model_paths
        self.models = []
        self.device = self._get_device()
        self.load_all_models()

    def _get_device(self):
        # Поддержка MPS (Mac), CUDA, fallback CPU
        if torch.cuda.is_available():
            return "cuda"
        try:
            if torch.backends.mps.is_available():
                return "mps"
        except Exception:
            pass
        return "cpu"

    def load_all_models(self):
        print("Loading models on device:", self.device)
        for p in self.model_paths:
            if not os.path.exists(p):
                print(f"Model not found: {p}")
                continue
            m = YOLO(p)
            try:
                m.to(self.device)
            except Exception:
                # ultralytics YOLO may manage device internally; ignore non-fatal
                pass
            self.models.append(m)
            print(f"Loaded {p} classes={len(m.names)} device={m.device}")

    def predict_ensemble_from_array(self, image_array: np.ndarray, conf=0.3, iou=0.6, model_weights: Optional[List[float]] = None):
        """
        image_array: BGR numpy array (cv2)
        возвращает: final_detections (list of dict), result_image (BGR numpy)
        """
        if len(self.models) < 1:
            raise RuntimeError("No models loaded")

        if model_weights is None:
            model_weights = [0.4, 0.3, 0.3]

        # Сохранение входного временного файла для передачи в модели, т.к. ultralytics принимает путь или array
        # ultralytics поддерживает np.ndarray input directly; используем это.
        all_detections = []  # [x1,y1,x2,y2,score,cls,model_idx]
        detection_counts = [0] * len(self.models)

        # Пройти по каждой модели
        for mi, model in enumerate(self.models):
            try:
                # ultralytics accepts np.ndarray; задаём conf через параметр
                results = model(image_array, conf=conf, verbose=False)
                for res in results:
                    if res.boxes is None:
                        continue
                    for box in res.boxes:
                        # box.xyxy is tensor shape (N,4). We access 0th because boxes iterable returns single per iteration
                        xyxy = box.xyxy[0].cpu().numpy() if hasattr(box.xyxy, "cpu") else box.xyxy[0].numpy()
                        x1, y1, x2, y2 = xyxy.tolist()
                        score = float(box.conf[0].cpu().numpy()) if hasattr(box.conf, "cpu") else float(box.conf[0].numpy())
                        cls = int(box.cls[0].cpu().numpy()) if hasattr(box.cls, "cpu") else int(box.cls[0].numpy())
                        all_detections.append([x1, y1, x2, y2, score, cls, mi])
                        detection_counts[mi] += 1
            except Exception as e:
                print(f"Error in model {mi}: {e}")

        print("Detections per model:", detection_counts)

        # Apply NMS across all detections
        final = self._apply_nms(all_detections, iou)
        if final is None or len(final) == 0:
            return [], image_array

        # final is numpy array of kept boxes; shape (k,6) if kept from code below
        # Convert final to structured list of dicts
        final_list = []
        for row in final:
            x1, y1, x2, y2, score, cls = row[:6]
            final_list.append({
                "x1": float(x1),
                "y1": float(y1),
                "x2": float(x2),
                "y2": float(y2),
                "score": float(score),
                "class": int(cls)
            })

        # Визуализация
        result_image = self.visualize_detections(image_array.copy(), final_list)
        return final_list, result_image

    def _apply_nms(self, detections, iou_threshold):
        if not detections:
            return None
        arr = np.array(detections)  # shape (N,7)
        boxes = torch.tensor(arr[:, :4], dtype=torch.float32)
        scores = torch.tensor(arr[:, 4], dtype=torch.float32)
        keep = nms(boxes, scores, iou_threshold)
        keep = keep.cpu().numpy()
        kept = arr[keep]
        # Normalize to have first 6 columns (x1,y1,x2,y2,score,cls)
        # some rows had model_idx at position 6; we drop it for final
        final = kept[:, :6]
        print(f"After NMS kept {final.shape[0]} boxes")
        return final

    def visualize_detections(self, image_bgr, detections):
        # Простейшая визуализация
        colors = [
            (0, 255, 0),
            (255, 0, 0),
            (0, 0, 255),
            (255, 255, 0),
            (255, 0, 255),
        ]
        class_names = ['scratch', 'dent', 'rust', 'other1', 'other2']

        for det in detections:
            x1, y1, x2, y2, score, cls = (det["x1"], det["y1"], det["x2"], det["y2"], det["score"], det["class"])
            color = colors[int(cls) % len(colors)]
            cv2.rectangle(image_bgr, (int(x1), int(y1)), (int(x2), int(y2)), color, 2)
            label = f"{class_names[int(cls)] if int(cls) < len(class_names) else 'Class'+str(int(cls))}: {score:.2f}"
            cv2.putText(image_bgr, label, (int(x1), max(15, int(y1)-5)), cv2.FONT_HERSHEY_SIMPLEX, 0.5, color, 1)
        return image_bgr

# ========== Инициализация ансамбля на старте ==========
ensemble = TripleEnsemble(MODEL_PATHS)

# ========== Эндпоинты ==========
@app.get("/health")
async def health():
    return {"status": "ok", "models_loaded2": len(ensemble.models)}

from fastapi.responses import FileResponse

@app.post("/analyze")
async def analyze_endpoint(file: UploadFile = File(...), conf: float = Form(0.3), iou: float = Form(0.6)):
    # Читаем файл
    content = await file.read()
    nparr = np.frombuffer(content, np.uint8)
    img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
    if img is None:
        return JSONResponse(status_code=400, content={"error": "invalid image"})

    try:
        detections, result_img = ensemble.predict_ensemble_from_array(img, conf=conf, iou=iou)
    except Exception as e:
        return JSONResponse(status_code=500, content={"error": str(e)})

    # Сохраняем изображение
    timestamp = int(time.time())
    out_name = f"ensemble_{timestamp}.jpg"
    out_path = os.path.join(RESULTS_DIR, out_name)
    cv2.imwrite(out_path, result_img)

    # Отдаём **фото напрямую**
    return FileResponse(out_path, media_type="image/jpeg", filename=out_name)


# ========== Запуск (если запуск напрямую) ==========
if __name__ == "__main__":
    import uvicorn
    uvicorn.run("app:app", host="0.0.0.0", port=5000, reload=False)
