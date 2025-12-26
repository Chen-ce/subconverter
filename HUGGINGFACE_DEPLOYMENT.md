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
CMD ["./subconverter", "serve", "--port", "7860", "--host", "0.0.0.0", "--log-level", "INFO"]
```

### 4. 设置 Secrets 与变量 (必须)
在 Hugging Face Space 的 **Settings** -> **Variables and secrets** 中添加：

**Secrets (加密保存):**
- `API_KEY`: 访问 Web 页面和管理配置所需的密钥。
- `HF_TOKEN`: 具有 **Write** 权限的 Hugging Face [Access Token](https://huggingface.co/settings/tokens)。

**Variables (公开变量):**
- `STORAGE_TYPE`: 设置为 `hf` 即可开启数据集持久化。
- `HF_REPO_ID`: 设置为你创建的私有数据集 ID (例如 `你的用户名/subconverter-data`)。

### 5. 数据持久化 (推荐方案)

由于 Hugging Face Spaces 默认文件系统重启即丢失，我们推荐使用私有 **Dataset** 进行持久化：

1. **创建数据集**: 在 Hugging Face 创建一个 [New Dataset](https://huggingface.co/new-dataset)，设置为 **Private**。
2. **配置变量**: 按照上述第 4 步，将 `STORAGE_TYPE` 设为 `hf`，并填入 `HF_REPO_ID` 和 `HF_TOKEN`。
3. **效果**: 生成的短链接 JSON 配置会自动实时同步到你的私有数据集中。即使 Space 重启或代码更新，数据也永不丢失。

### 6. 访问服务
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

## ❓ 常见问题
详见 [HF_TROUBLESHOOTING.md](HF_TROUBLESHOOTING.md)
