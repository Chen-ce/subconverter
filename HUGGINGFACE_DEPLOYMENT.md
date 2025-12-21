# Hugging Face Spaces 部署指南

## 前置准备

1. 注册 Hugging Face 账号：https://huggingface.co/join
2. 安装 Git（如果还没有）

## 部署步骤

### 1. 创建 Space

1. 访问 https://huggingface.co/new-space
2. 填写信息：
   - **Space name**: `subconverter`（或你喜欢的名字）
   - **License**: MIT
   - **Select the Space SDK**: Docker
   - **Space hardware**: CPU basic (免费)
3. 点击 **Create Space**

### 2. 克隆你的 Space 仓库

```bash
# 替换 YOUR_USERNAME 为你的 Hugging Face 用户名
git clone https://huggingface.co/spaces/YOUR_USERNAME/subconverter
cd subconverter
```

### 3. 复制项目文件

从你的 subconverter 项目复制以下文件到 Space 目录：

```bash
# 假设你的项目在 ~/subconverter
cp -r ~/Dev/Code/Go/subconverter/* .

# 或者手动复制这些文件/目录：
# - 所有 .go 文件
# - go.mod, go.sum
# - cmd/
# - config/
# - exporter/
# - fetcher/
# - merger/
# - parser/
# - server/
# - storage/
# - templates/
# - web/
# - config.yaml
# - Dockerfile (使用下面的 HF 专用版本)
```

### 4. 创建 Hugging Face 专用 Dockerfile

创建 `Dockerfile`（覆盖原有的）：

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o subconverter .

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制文件
COPY --from=builder /app/subconverter .
COPY --from=builder /app/web ./web
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/config.yaml .

# 创建数据目录
RUN mkdir -p /app/data/configs

# Hugging Face Spaces 使用 7860 端口
ENV PORT=7860
ENV HOST=0.0.0.0
ENV WEB_PATH=""

EXPOSE 7860

CMD ["./subconverter", "serve", "--port", "7860", "--host", "0.0.0.0"]
```

### 5. 创建 README.md

创建 `README.md`（Space 的说明页面）：

```markdown
---
title: Subconverter
emoji: 🔄
colorFrom: blue
colorTo: green
sdk: docker
pinned: false
---

# Subconverter - 订阅转换工具

一个功能强大的代理订阅合并和转换工具。

## 功能特性

- ✅ 多订阅合并
- ✅ 支持 12 种客户端格式
- ✅ ACL4SSR 规则库
- ✅ 短链接配置管理
- ✅ 现代化 Web 界面

## 使用方法

访问 Web 界面进行订阅转换和配置管理。

API 密钥请查看 Space 的 Secrets 设置。

## 项目地址

GitHub: https://github.com/Chen-ce/subconverter
```

### 6. 设置环境变量（Secrets）

1. 在 Space 页面点击 **Settings**
2. 找到 **Repository secrets**
3. 添加密钥：
   - Name: `API_KEY`
   - Value: 你的 API 密钥（使用 `openssl rand -hex 32` 生成）

### 7. 修改 config.yaml

确保 `config.yaml` 适配 Hugging Face：

```yaml
server:
  port: 7860  # Hugging Face 默认端口
  host: 0.0.0.0
  web_path: ""  # 根路径访问

auth:
  enabled: true
  api_keys:
    - ${API_KEY}  # 从环境变量读取

templates:
  dir: ./templates
  default: default

clash:
  default_rules: default
```

### 8. 推送到 Hugging Face

```bash
# 添加所有文件
git add .

# 提交
git commit -m "Initial deployment"

# 推送到 Hugging Face
git push
```

### 9. 等待构建

1. 推送后，Hugging Face 会自动开始构建 Docker 镜像
2. 在 Space 页面可以看到构建日志
3. 构建成功后，Space 会自动启动

### 10. 访问你的 Space

```
https://YOUR_USERNAME-subconverter.hf.space/
```

## 更新部署

当你需要更新代码时：

```bash
# 在 Space 目录中
git pull  # 如果有远程更新

# 复制新代码
cp -r ~/Dev/Code/Go/subconverter/* .

# 提交并推送
git add .
git commit -m "Update to latest version"
git push
```

## 常见问题

### Q: 构建失败怎么办？

A: 检查构建日志，常见问题：
- Go 版本不匹配
- 依赖下载失败
- Dockerfile 路径错误

### Q: 如何查看运行日志？

A: 在 Space 页面点击 **Logs** 标签

### Q: 如何更改 API 密钥？

A: 在 Settings → Repository secrets 中修改 `API_KEY`，然后重启 Space

### Q: Space 休眠了怎么办？

A: 免费 Space 会在不活动后休眠，访问 URL 会自动唤醒（可能需要等待几秒）

### Q: 如何使用自定义域名？

A: Hugging Face Spaces 免费版不支持自定义域名，需要升级到 Pro

## 性能优化

### 使用持久化存储

Hugging Face Spaces 支持持久化存储：

1. 在 Settings 中启用 **Persistent storage**
2. 修改 Dockerfile，将数据目录挂载到持久化路径：

```dockerfile
# 在 Dockerfile 中
ENV DATA_DIR=/data
RUN mkdir -p /data/configs
```

### 升级硬件

如果需要更好的性能，可以升级到：
- CPU upgrade: 更快的 CPU
- GPU: 如果需要 GPU 加速（本项目不需要）

## 故障排查

### 检查服务状态

访问：`https://YOUR_USERNAME-subconverter.hf.space/health.html`

### 查看日志

```bash
# 在 Space 页面的 Logs 标签查看实时日志
```

### 重启 Space

1. 在 Space 页面点击 **Settings**
2. 点击 **Factory reboot**

## 安全建议

1. **不要在代码中硬编码 API 密钥**
2. **使用 Secrets 管理敏感信息**
3. **定期更换 API 密钥**
4. **监控访问日志**

## 限制

Hugging Face Spaces 免费版限制：
- CPU: 2 vCPU
- RAM: 16 GB
- 存储: 50 GB（持久化需要额外配置）
- 休眠: 48 小时不活动后休眠

## 相关链接

- Hugging Face Spaces 文档: https://huggingface.co/docs/hub/spaces
- Docker SDK 文档: https://huggingface.co/docs/hub/spaces-sdks-docker
