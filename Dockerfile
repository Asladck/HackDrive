# Используем официальный Python 3.11 образ
FROM python:3.11-slim

# Устанавливаем зависимости системы
RUN apt-get update && apt-get install -y \
    build-essential \
    libgl1 \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы проекта
COPY ./app /app
COPY requirements.txt /app/requirements.txt

# Создаем папку для результатов
RUN mkdir -p /app/results

# Устанавливаем Python зависимости
RUN pip install --upgrade pip
RUN pip install --no-cache-dir -r requirements.txt

# Открываем порт для uvicorn
EXPOSE 5000

# Команда запуска
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "5000"]
