# check=skip=SecretsUsedInArgOrEnv
FROM node:22-alpine

WORKDIR /worker

ENV NODE_ENV=production
ENV IMAGE_WORKSPACE_API_BASE_URL=http://cloudbase:8080
ENV IMAGE_WORKSPACE_OUTPUT_DIR=/app/data/image-workspace
ENV IMAGE_WORKSPACE_STORAGE_KEY_ROOT=/app/data/image-workspace

COPY image-workspace-worker/package.json ./image-workspace-worker/
COPY image-workspace-worker/src ./image-workspace-worker/src

VOLUME ["/app/data"]

HEALTHCHECK --interval=30s --timeout=20s --start-period=20s --retries=3 \
  CMD node ./image-workspace-worker/src/worker.mjs --healthcheck

CMD ["node", "./image-workspace-worker/src/worker.mjs"]
