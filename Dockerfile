FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o subconverter .

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 复制二进制文件和配置
COPY --from=builder /app/subconverter .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/config.yaml .

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./subconverter", "serve"]
