# Hugging Face 部署故障排查

## 问题：无法加载 Web 界面 / 404 错误

### 1. 检查环境变量 `WEB_PATH`
**表现**：访问域名后显示 "404 page not found" 或页面样式丢失。
**原因**：在 Hugging Face Spaces 中，Space 域名通常直接映射到容器的根路径。如果你设置了 `WEB_PATH=/web`，你必须访问 `https://space-url/web/` 才能看到界面。
**建议**：
- 在 Space 的 **Settings** -> **Repository secrets** 中确保 `WEB_PATH` 留空。
- 或者在 `Dockerfile` 中定义 `ENV WEB_PATH=""`。

### 2. 静态文件未加载
**表现**：页面显示 HTML 但没有样式（style.css 404）或脚本不工作（app.js 404）。
**原因**：静态文件路径与 `WEB_PATH` 设置不匹配。
**解决**：确保 `config.yaml` 里的 `web_path` 与环境变量一致。本项目支持自动展开 `${WEB_PATH}`，但建议直接留空以适配根路径部署。

### 3. API Key (Token) 错误
**表现**：转换时报错 "请输入 API 密钥" 或 "身份验证失败"。
**解决**：
- 确保你已在 Space 的 **Settings** -> **Secrets** 中添加了 `API_KEY`。
- 确认你在 Web 界面输入的密钥与 Secret 里的 Value 完全一致。

### 4. 镜像拉取权限 (Unauthorized)
**表现**：Hugging Face 部署日志显示 `pull access denied`。
**原因**：GitHub Container Registry (ghcr.io) 镜像默认是私有的。
**解决**：
- 访问：`https://github.com/users/YOUR_USERNAME/packages/container/subconverter/settings`
- 将 **Visibility** 修改为 **Public**。

### 5. 数据丢失
**表现**：重启 Space 后，之前在配置页创建的正向/短链接配置消失了。
**原因**：Hugging Face Space 的容器文件系统是非持久化的。
**建议**：
- 在 Space Settings 中申请 **Persistent Storage** 并挂载到 `/app/data`。
- 或者将 `data` 目录下的 `config_storage.json` 导出备份。

---

## 调试命令

```bash
# 测试接口是否存活
curl -v https://YOUR_SPACE-subconverter.hf.space/health

# 测试订阅转换端点
curl -v https://YOUR_SPACE-subconverter.hf.space/api/sub

# 手动核对配置路径
# 如果你设置了 WEB_PATH=/foo, 那么你应该能访问：
# https://YOUR_SPACE-subconverter.hf.space/foo/index.html
```

如果仍有问题，请查看 Space 顶部的 **Logs** 标签获取 Go 服务运行时的具体日志。
