FROM node:22-alpine

RUN apk add --no-cache postgresql-client

WORKDIR /worker

ENV NODE_ENV=production
ENV HOT_WORKER_STATUS_PATH=/app/runtime/hot-worker-status.json
ENV HOT_RSS_COLLECT_INTERVAL_MS=1800000
ENV HOT_RSS_COLLECT_MAX_BACKOFF_MS=600000
ENV HOT_RSS_COLLECT_ON_START=true

COPY hot-rss-collector-worker.mjs ./

VOLUME ["/app/runtime"]

HEALTHCHECK --interval=30s --timeout=8s --start-period=20s --retries=3 \
  CMD node hot-rss-collector-worker.mjs --healthcheck

CMD ["node", "hot-rss-collector-worker.mjs"]
