# 构建前端
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# 构建后端
FROM golang:1.21-alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY backend/ ./
RUN go mod tidy && CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server

# 最终镜像
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=backend-builder /app/server .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

ENV TZ=Asia/Shanghai
ENV PORT=5555
ENV DB_PATH=/app/data/news.db
ENV DATA_DIR=/app/data

EXPOSE 5555

VOLUME ["/app/data"]

CMD ["./server"]
