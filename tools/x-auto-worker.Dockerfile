FROM python:3.12-slim

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    X_AUTO_PORT=8000 \
    X_ATUO_DATA_DIR=/app/data \
    X_ATUO_ENVIRONMENT=production \
    X_ATUO_SCHEDULER__ENABLED=false \
    X_ATUO_SCHEDULER__AUTOSTART=false

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY x-atuo/ /app/x-atuo/

RUN pip install --no-cache-dir /app/x-atuo

VOLUME ["/app/data"]

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD python -c "import os, urllib.request; urllib.request.urlopen('http://127.0.0.1:%s/healthz' % os.getenv('X_AUTO_PORT', '8000'), timeout=5).read()"

CMD ["sh", "-c", "uvicorn x_atuo.automation.api:app --host 0.0.0.0 --port ${X_AUTO_PORT:-8000}"]
