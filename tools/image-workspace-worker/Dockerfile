FROM node:22-alpine

WORKDIR /worker

ENV NODE_ENV=production
ENV IMAGE_WORKSPACE_API_BASE_URL=http://sub2api:8080
ENV IMAGE_WORKSPACE_OUTPUT_DIR=/app/data/image-workspace
ENV IMAGE_WORKSPACE_STORAGE_KEY_ROOT=/app/data/image-workspace

COPY package.json ./
COPY src ./src

VOLUME ["/app/data"]

HEALTHCHECK --interval=30s --timeout=8s --start-period=20s --retries=3 \
  CMD npm run healthcheck

CMD ["npm", "start"]
