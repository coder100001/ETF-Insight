FROM python:3.11-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

ENV ANALYTICS_SERVICE_PORT=8093
EXPOSE 8093

CMD ["python", "analytics_server.py"]
