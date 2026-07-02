# check=skip=SecretsUsedInArgOrEnv
FROM node:22-alpine

WORKDIR /worker

ENV NODE_ENV=production
ENV CLOUDBASE_BASE_URL=http://cloudbase:8080/api/v1
ENV WECHAT_EXPORT_OUTPUT_DIR=/app/data/wechat-export
ENV WECHAT_EXPORT_STORAGE_KEY_ROOT=/app/data/wechat-export

COPY wechat-worker/package.json wechat-worker/package-lock.json ./wechat-worker/
RUN npm --prefix ./wechat-worker ci --omit=dev

COPY wechat-worker/src ./wechat-worker/src

VOLUME ["/app/data"]

HEALTHCHECK --interval=30s --timeout=20s --start-period=20s --retries=3 \
  CMD npm --prefix ./wechat-worker run worker -- --healthcheck

CMD ["npm", "--prefix", "./wechat-worker", "run", "worker"]
