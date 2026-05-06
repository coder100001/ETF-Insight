FROM python:3.11-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

ENV DATA_SERVICE_PORT=8092
EXPOSE 8092

CMD ["python", "data_server.py"]
