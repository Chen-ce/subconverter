# Subconverter

一个功能强大的 Go 语言代理订阅合并和转换工具，支持 HTTP API 服务器和 Docker 部署。

## ✨ 功能特性

### 核心功能
- ✅ **合并多个订阅链接**：支持同时合并多个订阅源
- ✅ **支持单独添加节点**：可以添加独立的节点 URI
- ✅ **自动去重**：基于服务器地址和端口自动去除重复节点
- ✅ **节点过滤**：支持 include/exclude 关键词过滤
- ✅ **多协议支持**：VMess、VLESS、Trojan、Shadowsocks、ShadowsocksR
- ✅ **多格式支持**：Base64、Clash YAML

### 高级功能
- 🚀 **HTTP API 服务器**：提供 RESTful API 接口
- 🔐 **API 密钥认证**：支持 Bearer Token 和 URL 参数认证
- 📝 **自定义规则模板**：内置多个 Clash 规则模板（default、acl4ssr）
- 🐳 **Docker 部署**：完整的 Docker 和 docker-compose 支持
- 🎯 **智能代理组**：自动按地区分组（香港、台湾、新加坡、日本、美国、韩国）

## 📦 安装

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/guchenxi/subconverter.git
cd subconverter

# 下载依赖
go mod download

# 编译
go build -o subconverter
```

### 使用 Docker

```bash
# 使用 docker-compose
docker-compose up -d

# 或使用 Docker
docker build -t subconverter .
docker run -d -p 8080:8080 -e API_KEY=your-secret-key subconverter
```

## 🚀 快速开始

### 命令行模式（无需认证）

```bash
# 合并多个订阅
./subconverter merge \
  --sub "https://example.com/sub1" \
  --sub "https://example.com/sub2" \
  --output clash \
  --file output.yaml

# 添加单独节点
./subconverter merge \
  --node "vmess://..." \
  --node "trojan://..." \
  --output base64

# 过滤节点
./subconverter merge \
  --sub "https://example.com/sub1" \
  --include "香港|HK" \
  --exclude "x2|x3" \
  --output clash \
  --config acl4ssr
```

### HTTP API 模式（需要认证）

#### 1. 启动服务器

```bash
# 设置 API 密钥
export API_KEY=your-secret-api-key

# 启动服务器
./subconverter serve
```

#### 2. 调用 API

```bash
# 使用 Bearer Token 认证
curl "http://localhost:8080/sub?target=clash&url=https%3A%2F%2Fexample.com%2Fsub1" \
  -H "Authorization: Bearer your-secret-api-key"

# 使用 URL 参数认证
curl "http://localhost:8080/sub?target=clash&url=https%3A%2F%2Fexample.com%2Fsub1&token=your-secret-api-key"

# 合并多个订阅（用 | 分隔）
curl "http://localhost:8080/sub?target=clash&url=https%3A%2F%2Fexample.com%2Fsub1%7Chttps%3A%2F%2Fexample.com%2Fsub2&token=your-key"

# 过滤节点
curl "http://localhost:8080/sub?target=clash&url=...&include=香港|HK&exclude=x2&token=your-key"

# 使用自定义规则模板
curl "http://localhost:8080/sub?target=clash&url=...&config=acl4ssr&token=your-key"
```

## 📖 详细文档

### 命令行参数

#### `merge` 命令

| 参数 | 简写 | 说明 | 示例 |
|------|------|------|------|
| `--sub` | `-s` | 订阅链接（可多次指定） | `--sub "https://..."` |
| `--node` | `-n` | 单独节点 URI（可多次指定） | `--node "vmess://..."` |
| `--output` | `-o` | 输出格式（base64/clash） | `--output clash` |
| `--file` | `-f` | 输出文件路径 | `--file output.yaml` |
| `--include` | | 包含关键词（\| 分隔） | `--include "香港\|HK"` |
| `--exclude` | | 排除关键词（\| 分隔） | `--exclude "x2\|x3"` |
| `--config` | `-c` | Clash 规则模板名称 | `--config acl4ssr` |

#### `serve` 命令

| 参数 | 简写 | 说明 | 示例 |
|------|------|------|------|
| `--port` | `-p` | 服务器端口 | `--port 9090` |
| `--host` | `-H` | 服务器地址 | `--host 0.0.0.0` |
| `--config` | `-c` | 配置文件路径 | `--config config.yaml` |

### API 端点

#### `GET /health`
健康检查端点（无需认证）

**响应示例：**
```json
{
  "status": "ok",
  "message": "subconverter is running"
}
```

#### `GET /sub`
订阅转换端点（需要认证）

**参数：**
- `target`: 输出格式（clash, base64）
- `url`: 订阅链接（多个用 | 分隔，需 URL 编码）
- `node`: 单独节点（可多个）
- `config`: 规则模板名称（可选，默认 default）
- `include`: 包含关键词（可选，| 分隔）
- `exclude`: 排除关键词（可选，| 分隔）
- `token`: API 密钥（可选，也可用 Header）

**认证方式：**
1. Header: `Authorization: Bearer <API_KEY>`
2. URL 参数: `?token=<API_KEY>`

#### `GET /templates`
获取可用模板列表（需要认证）

**响应示例：**
```json
{
  "templates": ["default", "acl4ssr"],
  "default": "default"
}
```

### 规则模板

#### `default` 模板
基础规则模板，包含：
- 🚀 节点选择
- ♻️ 自动选择
- 🎯 全球直连
- 常用网站规则
- GeoIP 中国直连

#### `acl4ssr` 模板
ACL4SSR 风格规则模板，包含：
- 🚀 节点选择
- ♻️ 自动选择
- 🇭🇰 香港节点
- 🇨🇳 台湾节点
- 🇸🇬 狮城节点
- 🇯🇵 日本节点
- 🇺🇲 美国节点
- 🇰🇷 韩国节点
- 🛑 广告拦截
- 🐟 漏网之鱼
- 完整的分流规则

## 🐳 Docker 部署

### 使用 docker-compose（推荐）

1. **创建 `.env` 文件**

```bash
API_KEY=your-secret-api-key-change-this
PORT=8080
```

2. **启动服务**

```bash
docker-compose up -d
```

3. **查看日志**

```bash
docker-compose logs -f
```

4. **停止服务**

```bash
docker-compose down
```

### 使用 Docker

```bash
# 构建镜像
docker build -t subconverter .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -e API_KEY=your-secret-key \
  -v $(pwd)/config.yaml:/root/config.yaml \
  -v $(pwd)/templates:/root/templates \
  --name subconverter \
  subconverter

# 查看日志
docker logs -f subconverter

# 停止容器
docker stop subconverter
docker rm subconverter
```

## ⚙️ 配置文件

### `config.yaml`

```yaml
server:
  port: 8080
  host: 0.0.0.0

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

### 环境变量

- `API_KEY`: API 密钥（必需）
- `PORT`: 服务器端口（可选，默认 8080）
- `HOST`: 服务器地址（可选，默认 0.0.0.0）

## 🔒 安全建议

> [!WARNING]
> **生产环境安全建议**

1. **使用强密钥**：API_KEY 应使用至少 32 位随机字符串
   ```bash
   # 生成随机密钥
   openssl rand -hex 32
   ```

2. **启用 HTTPS**：使用 Nginx 或 Caddy 作为反向代理
   ```nginx
   server {
       listen 443 ssl;
       server_name sub.example.com;
       
       ssl_certificate /path/to/cert.pem;
       ssl_certificate_key /path/to/key.pem;
       
       location / {
           proxy_pass http://localhost:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
       }
   }
   ```

3. **限流保护**：建议在反向代理层添加 rate limiting

4. **日志监控**：定期检查访问日志，监控异常访问

## 📝 使用示例

### 示例 1: 合并三个订阅并使用 ACL4SSR 规则

```bash
./subconverter merge \
  --sub "https://example1.com/subscription" \
  --sub "https://example2.com/subscription" \
  --sub "https://example3.com/subscription" \
  --output clash \
  --config acl4ssr \
  --file merged.yaml
```

### 示例 2: 只保留香港和新加坡节点

```bash
./subconverter merge \
  --sub "https://example.com/subscription" \
  --include "香港|HK|新加坡|SG" \
  --output clash \
  --file hk_sg_only.yaml
```

### 示例 3: 排除高倍率节点

```bash
./subconverter merge \
  --sub "https://example.com/subscription" \
  --exclude "x2|x3|x5" \
  --output clash
```

### 示例 4: API 调用示例（JavaScript）

```javascript
const API_KEY = 'your-secret-key';
const subscriptions = [
  'https://example1.com/sub',
  'https://example2.com/sub'
];

const url = new URL('http://localhost:8080/sub');
url.searchParams.set('target', 'clash');
url.searchParams.set('url', subscriptions.join('|'));
url.searchParams.set('config', 'acl4ssr');
url.searchParams.set('token', API_KEY);

fetch(url)
  .then(res => res.text())
  .then(config => {
    console.log('Clash config:', config);
  });
```

## 🛠️ 开发

```bash
# 运行测试
go test ./...

# 格式化代码
go fmt ./...

# 运行 linter
golangci-lint run
```

## 📂 项目结构

```
subconverter/
├── cmd/                    # CLI 命令
│   ├── root.go            # 根命令
│   ├── merge.go           # 合并命令
│   └── serve.go           # 服务器命令
├── parser/                # 订阅解析器
│   ├── types.go           # 通用类型
│   ├── base64.go          # Base64 解析器
│   ├── clash.go           # Clash 解析器
│   └── node.go            # 节点 URI 解析器
├── fetcher/               # 订阅获取器
│   └── fetcher.go         # HTTP 获取器
├── merger/                # 节点合并器
│   └── merger.go          # 合并和去重
├── exporter/              # 格式导出器
│   ├── base64.go          # Base64 导出器
│   ├── clash.go           # Clash 导出器
│   └── clash_custom.go    # 自定义配置导出
├── server/                # HTTP 服务器
│   ├── server.go          # 服务器主逻辑
│   ├── middleware/        # 中间件
│   │   ├── auth.go        # 认证中间件
│   │   └── cors.go        # CORS 中间件
│   └── handler/           # 请求处理器
│       ├── convert.go     # 转换处理器
│       ├── health.go      # 健康检查
│       └── templates.go   # 模板列表
├── config/                # 配置管理
│   └── config.go          # 配置加载和验证
├── templates/             # 规则模板
│   └── rules/             # Clash 规则模板
│       ├── default.yaml   # 默认模板
│       └── acl4ssr.yaml   # ACL4SSR 模板
├── config.yaml            # 配置文件
├── Dockerfile             # Docker 镜像
├── docker-compose.yml     # Docker Compose 配置
└── main.go                # 程序入口
```

## 📄 License

MIT

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！
