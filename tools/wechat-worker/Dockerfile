FROM node:22-alpine

WORKDIR /worker

ENV NODE_ENV=production
ENV SUB2API_BASE_URL=http://sub2api:8080/api/v1
ENV WECHAT_EXPORT_OUTPUT_DIR=/app/data/wechat-export
ENV WECHAT_EXPORT_STORAGE_KEY_ROOT=/app/data/wechat-export

COPY package.json package-lock.json ./
RUN npm ci --omit=dev

COPY src ./src

VOLUME ["/app/data"]

HEALTHCHECK --interval=30s --timeout=8s --start-period=20s --retries=3 \
  CMD npm run worker -- --healthcheck

CMD ["npm", "run", "worker"]
