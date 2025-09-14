# Используем легковесный Python 3.11
FROM python:3.11-slim

# Устанавливаем зависимости ОС для OpenCV
RUN apt-get update && apt-get install -y \
    build-essential \
    libgl1 \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

# Рабочая директория в контейнере
WORKDIR /app

# Копируем файлы проекта
COPY ./app /app
COPY requirements.txt /app/requirements.txt

# Устанавливаем зависимости Python
RUN pip install --no-cache-dir --upgrade pip
RUN pip install --no-cache-dir -r /app/requirements.txt

# Экспонируем порт FastAPI
EXPOSE 5000

# Команда запуска
CMD ["uvicorn", "cmd.main:app", "--host", "0.0.0.0", "--port", "5000"]
