FROM node:22-alpine

WORKDIR /worker

ENV NODE_ENV=production
ENV SUB2API_BASE_URL=http://sub2api:8080/api/v1
ENV IMAGE_WORKSPACE_API_BASE_URL=http://sub2api:8080
ENV WECHAT_EXPORT_OUTPUT_DIR=/app/data/wechat-export
ENV WECHAT_EXPORT_STORAGE_KEY_ROOT=/app/data/wechat-export
ENV IMAGE_WORKSPACE_OUTPUT_DIR=/app/data/image-workspace
ENV IMAGE_WORKSPACE_STORAGE_KEY_ROOT=/app/data/image-workspace

COPY wechat-worker/package.json wechat-worker/package-lock.json ./wechat-worker/
RUN npm --prefix ./wechat-worker ci --omit=dev

COPY wechat-worker/src ./wechat-worker/src
COPY image-workspace-worker/package.json ./image-workspace-worker/
COPY image-workspace-worker/src ./image-workspace-worker/src
COPY business-worker.mjs ./

VOLUME ["/app/data"]

HEALTHCHECK --interval=30s --timeout=20s --start-period=20s --retries=3 \
  CMD node business-worker.mjs --healthcheck

CMD ["node", "business-worker.mjs"]
