# Subconverter

一个功能强大的 Go 语言代理订阅合并和转换工具，支持多种客户端格式、短链接配置管理和 Docker 部署。

## ✨ 核心功能

### 🎯 订阅转换
- **多订阅合并**：同时合并多个订阅源
- **单独节点添加**：支持添加独立节点 URI
- **智能去重**：基于协议+凭证的精确去重
- **节点过滤**：include/exclude 关键词过滤

### 📱 全客户端支持
- **Clash** - 完整支持，含规则模板
- **Sing-box** - JSON 格式
- **Surge 2/3/4** - 所有版本
- **Surfboard** - 基于 Surge 4
- **Quantumult X** - 完整支持
- **Loon** - 完整支持
- **V2Ray/Shadowsocks/SSR** - Base64 URI

### 🔗 短链接配置系统
- **配置持久化**：JSON 文件存储
- **随时修改**：更新配置无需更改订阅链接
- **Web 管理界面**：可视化配置管理
- **安全隐私**：订阅链接不暴露在 URL 中

### 📚 ACL4SSR 规则库
- **在线基础版**：去广告、自动测速
- **在线完整版**：流媒体、AI、游戏分组
- **在线精简版**：轻量级规则
- **强化去广告版**：增强拦截
- **无自动测速版**：手动控制

## 🚀 快速开始

### Docker 部署（推荐）

```bash
# 1. 创建 .env 文件
echo "API_KEY=$(openssl rand -hex 32)" > .env

# 2. 启动服务
docker-compose up -d

# 3. 访问 Web 界面
open http://localhost:8080/web
```

### 从源码编译

```bash
git clone https://github.com/Chen-ce/subconverter.git
cd subconverter
go build -o subconverter
./subconverter serve
```

## 📖 使用方式

### 方式 1: Web 界面（最简单）

访问 `http://localhost:8080/web`

**转换工具**：
1. 输入 API 密钥
2. 添加订阅链接
3. 选择客户端类型和规则模板
4. 点击转换，获取订阅链接

**配置管理**：
1. 创建配置（输入订阅、节点、规则）
2. 获得短链接（如 `http://server/sub/abc123`）
3. 添加到 Clash/Surge 等客户端
4. 随时编辑配置，客户端自动同步

### 方式 2: API 调用

```bash
# 基础转换
curl "http://localhost:8080/sub?target=clash&url=订阅链接&token=your-api-key"

# 使用 ACL4SSR 完整版
curl "http://localhost:8080/sub?target=clash&url=订阅链接&config=acl4ssr_online_full&token=xxx"

# Surge 4
curl "http://localhost:8080/sub?target=surge&ver=4&url=订阅链接&token=xxx"

# Sing-box
curl "http://localhost:8080/sub?target=singbox&url=订阅链接&token=xxx"

# 创建短链接配置
curl -X POST http://localhost:8080/api/config \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "urls": ["订阅1", "订阅2"],
    "target": "clash",
    "config": "acl4ssr_online_full"
  }'
# 返回: {"id": "abc123", "url": "http://server/sub/abc123"}

# 使用短链接
curl "http://localhost:8080/sub/abc123?token=your-api-key"
```

### 方式 3: 命令行

```bash
# 合并订阅
./subconverter merge \
  --sub "订阅1" \
  --sub "订阅2" \
  --output clash \
  --config acl4ssr \
  --file output.yaml

# 过滤节点
./subconverter merge \
  --sub "订阅链接" \
  --include "香港|日本" \
  --exclude "x2|x3" \
  --output clash
```

## 📚 规则模板说明

| 模板 | 特点 | 适用场景 |
|------|------|---------|
| **acl4ssr_online** | ✅ 去广告 ✅ 自动测速 ✅ 基础分流 | 日常使用 |
| **acl4ssr_online_full** | ✅ 流媒体 ✅ AI服务 ✅ 游戏平台 | 节点丰富 |
| **acl4ssr_online_mini** | ⚡ 轻量级 ✅ 核心规则 | 节点较少 |
| **acl4ssr_online_adblock** | 🛡️ 强化去广告 ✅ 隐私保护 | 注重去广告 |
| **acl4ssr_online_noauto** | ❌ 无自动测速 ✅ 手动选择 | 手动控制 |

详见：[ACL4SSR_TEMPLATES.md](ACL4SSR_TEMPLATES.md)

## 🔧 API 端点

| 端点 | 方法 | 说明 | 认证 |
|------|------|------|------|
| `/health` | GET | 健康检查 | ❌ |
| `/sub` | GET | 订阅转换 | ✅ |
| `/sub/:id` | GET | 短链接订阅 | ✅ |
| `/templates` | GET | 模板列表 | ✅ |
| `/api/config` | POST | 创建配置 | ✅ |
| `/api/config/:id` | GET | 查看配置 | ✅ |
| `/api/config/:id` | PUT | 更新配置 | ✅ |
| `/api/config/:id` | DELETE | 删除配置 | ✅ |
| `/api/configs` | GET | 配置列表 | ✅ |

### 订阅转换参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `target` | 客户端类型 | `clash`, `singbox`, `surge`, `quanx`, `loon` |
| `url` | 订阅链接（\| 分隔） | `url1\|url2` |
| `node` | 单独节点（可多个） | `vmess://...` |
| `config` | 规则模板 | `acl4ssr_online_full` |
| `include` | 包含关键词 | `香港\|日本` |
| `exclude` | 排除关键词 | `x2\|x3` |
| `ver` | Surge 版本 | `2`, `3`, `4` |
| `token` | API 密钥 | - |

## 🐳 Docker 部署

### docker-compose（推荐）

```yaml
version: '3'
services:
  subconverter:
    image: subconverter:latest
    ports:
      - "8080:8080"
    environment:
      - API_KEY=${API_KEY}
    volumes:
      - ./config.yaml:/app/config.yaml
      - ./templates:/app/templates
      - ./data:/app/data  # 配置持久化
    restart: unless-stopped
```

### 环境变量

```bash
# .env
API_KEY=your-secret-api-key-change-this
PORT=8080
```

## 🔒 安全建议

> [!WARNING]
> **生产环境必读**

1. **强密钥**：使用 32 位以上随机字符串
   ```bash
   openssl rand -hex 32
   ```

2. **HTTPS**：使用 Nginx/Caddy 反向代理
   ```nginx
   server {
       listen 443 ssl;
       server_name sub.example.com;
       location / {
           proxy_pass http://localhost:8080;
       }
   }
   ```

3. **限流**：防止 API 滥用
4. **监控**：定期检查访问日志

## 📂 项目结构

```
subconverter/
├── cmd/                    # CLI 命令
├── parser/                 # 订阅解析器
├── fetcher/                # 订阅获取器
├── merger/                 # 节点合并器（精确去重）
├── exporter/               # 格式导出器
│   ├── clash.go           # Clash
│   ├── singbox.go         # Sing-box
│   ├── surge.go           # Surge
│   ├── quantumultx.go     # Quantumult X
│   ├── loon.go            # Loon
│   └── base64.go          # Base64
├── server/                 # HTTP 服务器
│   ├── middleware/        # 认证、CORS
│   └── handler/           # API 处理器
│       ├── convert.go     # 订阅转换
│       └── config.go      # 配置管理
├── storage/                # 配置存储（JSON）
├── templates/              # 规则模板
│   ├── template.go        # 模板管理器
│   └── rules/             # Clash 规则
├── web/                    # Web 界面
│   ├── index.html         # 转换工具
│   ├── configs.html       # 配置管理
│   ├── app.js
│   ├── configs.js
│   └── style.css
├── config.yaml             # 配置文件
├── docker-compose.yml
└── Dockerfile
```

## 🎯 使用场景

### 场景 1: 多机场合并
```
机场1 + 机场2 + 机场3 → 一个订阅链接
✅ 自动去重
✅ 统一规则
✅ 自动更新
```

### 场景 2: 自建节点混用
```
机场订阅 + 自建 VPS → 统一管理
✅ 保留所有节点
✅ 灵活分组
```

### 场景 3: 团队共享
```
创建配置 → 生成短链接 → 分享给团队
✅ 统一配置
✅ 集中管理
✅ 随时更新
```

## 📄 License

MIT

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 🔗 相关链接

- [ACL4SSR 规则库](https://github.com/ACL4SSR/ACL4SSR)
- [Clash 文档](https://github.com/Dreamacro/clash)
- [Sing-box 文档](https://sing-box.sagernet.org/)
