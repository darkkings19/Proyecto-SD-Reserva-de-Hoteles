FROM python:3.11-slim

WORKDIR /app

# Configurar variables de entorno para Python
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PYTHONPATH=/app/src

# Instalar dependencias
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copiar el código fuente
COPY src/ ./src/

# Exponer el puerto por defecto (se puede sobreescribir con docker-compose)
EXPOSE 50051

# Comando para ejecutar el servidor
CMD ["python", "src/main.py"]
