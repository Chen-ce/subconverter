# 使用 GitHub Actions + 预构建镜像部署到 Hugging Face

这种方式更高效，GitHub Actions 自动构建镜像，Hugging Face 直接使用。

## 优势

✅ **更快的部署** - Hugging Face 不需要重新编译  
✅ **自动化** - 推送代码自动构建镜像  
✅ **多平台支持** - 同时构建 amd64 和 arm64  
✅ **缓存优化** - GitHub Actions 缓存加速构建  

## 设置步骤

### 1. 配置 GitHub Secrets（可选 Docker Hub）

如果要推送到 Docker Hub，在 GitHub 仓库设置中添加：

1. 访问 `https://github.com/Chen-ce/subconverter/settings/secrets/actions`
2. 添加 Secrets：
   - `DOCKERHUB_USERNAME`: 你的 Docker Hub 用户名
   - `DOCKERHUB_TOKEN`: Docker Hub Access Token

**获取 Docker Hub Token：**
1. 访问 https://hub.docker.com/settings/security
2. 点击 **New Access Token**
3. 复制生成的 token

### 2. 推送代码触发构建

```bash
git add .github/workflows/docker-build.yml
git commit -m "Add GitHub Actions workflow for Docker build"
git push
```

GitHub Actions 会自动：
- 构建 Docker 镜像
- 推送到 GitHub Container Registry (ghcr.io)
- 推送到 Docker Hub（如果配置了）

### 3. 查看构建状态

访问 `https://github.com/Chen-ce/subconverter/actions`

### 4. 部署到 Hugging Face

#### 方法 A: 使用 GitHub Container Registry（推荐，无需额外配置）

1. 创建 Hugging Face Space（SDK: Docker）
2. 克隆 Space 仓库：
   ```bash
   git clone https://huggingface.co/spaces/YOUR_USERNAME/subconverter
   cd subconverter
   ```

3. 创建 `Dockerfile`：
   ```dockerfile
   FROM ghcr.io/chen-ce/subconverter:latest
   
   ENV PORT=7860
   ENV HOST=0.0.0.0
   ENV WEB_PATH=""
   
   EXPOSE 7860
   
   CMD ["./subconverter", "serve", "--port", "7860", "--host", "0.0.0.0"]
   ```

4. 创建 `README.md`（复制 `README_HF.md` 的内容）

5. 推送到 Hugging Face：
   ```bash
   git add Dockerfile README.md
   git commit -m "Deploy using pre-built image"
   git push
   ```

#### 方法 B: 使用 Docker Hub

如果你配置了 Docker Hub，Dockerfile 改为：

```dockerfile
FROM YOUR_DOCKERHUB_USERNAME/subconverter:latest

ENV PORT=7860
ENV HOST=0.0.0.0
ENV WEB_PATH=""

EXPOSE 7860

CMD ["./subconverter", "serve", "--port", "7860", "--host", "0.0.0.0"]
```

### 5. 设置 API 密钥

在 Hugging Face Space Settings → Repository secrets：
- Name: `API_KEY`
- Value: 你的 API 密钥

### 6. 访问你的 Space

```
https://YOUR_USERNAME-subconverter.hf.space/
```

## 更新流程

当你更新代码时：

```bash
# 1. 推送到 GitHub
git add .
git commit -m "Update features"
git push

# 2. GitHub Actions 自动构建新镜像

# 3. 在 Hugging Face Space 中触发重新部署
# 访问 Space Settings → Factory reboot
# 或者推送一个空提交：
cd huggingface-space
git commit --allow-empty -m "Trigger rebuild"
git push
```

## 镜像标签

GitHub Actions 会创建多个标签：

- `latest` - 最新的 main 分支构建
- `main` - main 分支
- `sha-xxxxxx` - 特定 commit
- `v1.0.0` - 版本标签（如果打了 tag）

在 Hugging Face 中可以指定具体版本：

```dockerfile
# 使用特定版本
FROM ghcr.io/chen-ce/subconverter:v1.0.0

# 使用特定 commit
FROM ghcr.io/chen-ce/subconverter:sha-abc1234
```

## 故障排查

### GitHub Actions 构建失败

1. 检查 Actions 日志
2. 确保 Dockerfile 正确
3. 检查 Go 依赖是否完整

### Hugging Face 拉取镜像失败

**GitHub Container Registry 是私有的？**

默认情况下，GitHub Container Registry 的镜像是私有的。需要设置为公开：

1. 访问 `https://github.com/users/Chen-ce/packages/container/subconverter/settings`
2. 在 **Danger Zone** 中点击 **Change visibility**
3. 选择 **Public**

或者在 GitHub Actions 中添加权限设置（已包含在 workflow 中）。

### 查看可用镜像

**GitHub Container Registry:**
```bash
# 查看所有标签
curl https://ghcr.io/v2/chen-ce/subconverter/tags/list
```

**Docker Hub:**
```bash
# 访问
https://hub.docker.com/r/YOUR_USERNAME/subconverter/tags
```

## 性能对比

| 方式 | 构建时间 | 部署时间 | 总时间 |
|------|---------|---------|--------|
| **直接构建** | - | ~5-10分钟 | ~5-10分钟 |
| **预构建镜像** | ~3-5分钟 | ~30秒 | ~3.5-5.5分钟 |

使用预构建镜像，Hugging Face 部署时间从 5-10 分钟缩短到 30 秒！

## 高级配置

### 多阶段构建优化

当前 Dockerfile 已经使用多阶段构建，最终镜像大小约 20-30 MB。

### 自动更新

可以设置 GitHub Actions 定时触发：

```yaml
on:
  schedule:
    - cron: '0 0 * * 0'  # 每周日构建一次
```

### 构建通知

添加构建状态徽章到 README：

```markdown
![Docker Build](https://github.com/Chen-ce/subconverter/actions/workflows/docker-build.yml/badge.svg)
```

## 相关链接

- [GitHub Container Registry 文档](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Docker Hub](https://hub.docker.com/)
- [Hugging Face Spaces Docker SDK](https://huggingface.co/docs/hub/spaces-sdks-docker)
