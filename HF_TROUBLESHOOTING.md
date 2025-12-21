# Hugging Face 部署故障排查

## 问题：访问根路径返回 404

### 可能原因

1. **镜像未公开** - GitHub Container Registry 镜像默认是私有的
2. **使用旧镜像** - HF 缓存了旧版本镜像
3. **路由问题** - 代码路由配置有误

### 解决步骤

#### 1. 设置镜像为公开（重要！）

访问：https://github.com/users/Chen-ce/packages/container/subconverter/settings

在 **Danger Zone** 中：
1. 点击 **Change visibility**
2. 选择 **Public**
3. 输入仓库名确认

#### 2. 验证镜像可访问

```bash
# 尝试拉取镜像
docker pull ghcr.io/chen-ce/subconverter:latest

# 如果失败，说明镜像是私有的或不存在
```

#### 3. 检查 HF Space 日志

1. 访问：https://huggingface.co/spaces/chence0918/subconverter
2. 点击 **Logs** 标签
3. 查看是否有错误信息

常见错误：
- `Error: pull access denied` - 镜像是私有的
- `404 page not found` - 路由问题
- `connection refused` - 服务未启动

#### 4. 强制重新拉取镜像

在 HF Space 中：

```bash
# 方法 A: 修改 Dockerfile 强制重新拉取
# 在 FROM 行添加特定标签
FROM ghcr.io/chen-ce/subconverter:sha-xxxxxx

# 方法 B: 空提交触发重建
git commit --allow-empty -m "Force rebuild"
git push
```

#### 5. 临时解决方案：使用完整构建

如果镜像访问有问题，可以暂时使用完整构建：

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o subconverter

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/subconverter .
COPY --from=builder /app/web ./web
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/config.yaml .

ENV PORT=7860
EXPOSE 7860
CMD ["./subconverter", "serve", "--port", "7860", "--host", "0.0.0.0"]
```

### 验证本地构建

```bash
# 1. 构建镜像
docker build -t subconverter-test .

# 2. 运行容器
docker run -p 7860:7860 -e API_KEY=test123 subconverter-test

# 3. 测试访问
curl http://localhost:7860/
curl http://localhost:7860/health
```

### 调试命令

```bash
# 查看 HF Space 实时日志
# 在 Space 页面的 Logs 标签

# 检查镜像标签
curl https://ghcr.io/v2/chen-ce/subconverter/tags/list

# 测试路由
curl -v https://chence0918-subconverter.hf.space/
curl -v https://chence0918-subconverter.hf.space/health
curl -v https://chence0918-subconverter.hf.space/api/sub
```

### 当前状态检查清单

- [ ] GitHub Actions 构建成功
- [ ] 镜像设置为公开
- [ ] HF Space Dockerfile 正确
- [ ] HF Space 成功拉取镜像
- [ ] 服务在容器内启动
- [ ] 根路径返回 200

### 联系支持

如果以上都无法解决，可以：
1. 在 HF Space 的 Community 标签发帖
2. 查看 HF Docs: https://huggingface.co/docs/hub/spaces-sdks-docker
