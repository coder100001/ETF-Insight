FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

ENV AGENT_SERVICE_PORT=8091
EXPOSE 8091

CMD ["python", "agent_server.py"]
