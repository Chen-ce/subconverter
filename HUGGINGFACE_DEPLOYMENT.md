# Hugging Face Spaces 部署指南

本项目原生支持 Hugging Face Spaces，使用 Docker SDK 部署，支持 GitHub Actions 自动构建并同步。

## 🚀 推荐部署方式：GitHub Actions + 预构建镜像

这种方式最稳定且更新最快。

### 1. 准备工作
- 在 Hugging Face 创建一个 **New Space**。
- 选择 **Docker** SDK。
- 此时你会得到一个空的仓库链接（如 `https://huggingface.co/spaces/YOUR_USERNAME/subconverter`）。

### 2. 配置仓库
将 `README_HF.md` 的内容复制到 Space 的 `README.md` 中。这会让 Hugging Face 识别它是 Docker 项目并渲染美观的介绍页面。

### 3. 创建 Dockerfile
在 Space 仓库中创建一个 `Dockerfile`：

```dockerfile
# 使用 GitHub Actions 自动构建的镜像
FROM ghcr.io/chen-ce/subconverter:latest

# 配置环境变量
ENV PORT=7860
ENV HOST=0.0.0.0
# 重要：确保 WEB_PATH 为空（根路径），以便 Space 域名直接访问
ENV WEB_PATH=""

EXPOSE 7860

# 启动命令
CMD ["./subconverter", "serve", "--port", "7860", "--host", "0.0.0.0"]
```

### 4. 设置 Secrets (必须)
在 Hugging Face Space 的 **Settings** -> **Repository secrets** 中添加：

- `API_KEY`: 访问 Web 页面和管理历史配置所需的密钥。建议使用 `openssl rand -hex 16` 生成。

### 5. 访问服务
稍等 1-2 分钟，构建完成后即可通过 Space 域名预览：
`https://YOUR_USERNAME-subconverter.hf.space/`

---

## 🛠️ 运维与更新

### 如何同步最新功能？
当我的主仓库（GitHub）有代码更新时，GitHub Actions 会自动构建新的镜像。
你只需要刷新你的 Space 或执行 **Factory reboot** 即可拉取最新代码：

1. 进入 Space 的 **Settings**。
2. 找到 **Danger Zone**。
3. 点击 **Factory reboot**。

### 如何持久化数据？
Hugging Face Spaces 的文件系统默认是**非持久化**的（虽然项目使用了 JSON 文件存储配置，但重启即丢）。
- 如果你需要永久保存短链接配置，建议使用 Hugging Face 提供的 **Persistent Storage** 云盘挂载到 `/app/data`。
- 或者定期备份 `data/config_storage.json`。

## ❓ 常见问题
详见 [HF_TROUBLESHOOTING.md](HF_TROUBLESHOOTING.md)
