FROM python:3.12-slim

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    HOT_WORKER_STATUS_PATH=/app/runtime/hot-worker-status.json \
    HOT_RSS_COLLECT_INTERVAL_MS=1800000 \
    HOT_RSS_COLLECT_MAX_BACKOFF_MS=600000 \
    HOT_RSS_COLLECT_ON_START=true

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY x-atuo/ /app/x-atuo/

RUN pip install --no-cache-dir /app/x-atuo

VOLUME ["/app/runtime"]

HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD python -m x_atuo.hot_rss_worker --healthcheck

CMD ["python", "-m", "x_atuo.hot_rss_worker"]
