---
title: Subconverter
emoji: 🔄
colorFrom: blue
colorTo: green
sdk: docker
pinned: false
---

# Subconverter

一个功能强大的 Go 语言代理订阅合并和转换工具，支持多种客户端格式、短链接配置管理和 Docker 部署。内置极简现代的 Web 界面，支持一键生成及持久化管理。

## ✨ 核心功能

### 🎯 订阅转换
- **多订阅合并**：同时合并多个订阅源，支持 URL | 分隔。
- **单独节点添加**：支持直接添加 VMess、VLESS、Trojan、SS、SSR 节点 URI。
- **智能去重**：基于协议、地址和端口的精确去重逻辑。
- **节点过滤**：支持 include/exclude 正则或关键词过滤。

### 📱 全客户端支持
- **Clash** - 支持规则模板、分组控制。
- **Sing-box** - 标准 JSON 格式导出。
- **Surge 2/3/4** - 自动处理版本差异参数。
- **Quantumult X / Loon** - 完整支持。
- **V2Ray/Shadowsocks/SSR** - 标准 Base64 链接。

### 🔗 短链接与配置编辑器
- **iOS 风格界面**：首页支持一键 Toggle 生成短链接（默认开启）。
- **内容预览**：转换后可直接预览节点内容，无需下载。
- **配置编辑器**：支持“反向解析”，粘贴转换链接即可恢复所有配置项。
- **历史管理**：配置页面支持按 API Key 自动加载所有历史记录，一键切换编辑。
- **持久化存储**：不依赖数据库，使用轻量级 JSON 文件持久化配置。

## 🚀 快速开始

### Docker 部署（推荐）

```bash
# 1. 创建 .env 文件
echo "API_KEY=$(openssl rand -hex 16)" > .env

# 2. 启动服务
docker-compose up -d

# 3. 访问 Web 界面
open http://localhost:8080/
```

### 从源码编译

```bash
git clone https://github.com/Chen-ce/subconverter.git
cd subconverter
go build -o subconverter
./subconverter serve
```

## 📖 使用指南

### Web 界面

默认访问地址：`http://localhost:8080/`

**1. 快速转换 (首页)**
- **API 密钥**：初次使用需输入 `.env` 中定义的密钥。
- **订阅链接**：每行一个订阅，或直接粘贴带参数的长链接（系统会自动识别）。
- **生成短链接**：默认开启。转换后你将得到一个极简的 `/sub/{id}` 链接，方便在 iOS/Mac 客户端上长期订阅。
- **预览内容**：转换成功后点击“预览内容”，可直接查看合并后的节点明文。

**2. 配置编辑 (管理页)**
- **历史列表**：输入密钥后，左边栏会显示你名下所有的短链接配置。
- **按图索骥**：点击列表项，所有订阅、节点、过滤规则将瞬间载入右侧编辑器。
- **反向还原**：粘贴一个之前的转换长链接，编辑器会自动将其“拆解”回原始参数。
- **即时更新**：修改后点击“保存配置”，你的订阅链接内容即刻同步，无需更换 URL。

### API 调用

```bash
# 基础转换 (Clash)
curl "http://localhost:8080/sub?target=clash&url=订阅链接&token=xxx"

# 带过滤规则
curl "http://localhost:8080/sub?target=clash&url=URL1|URL2&include=香港&exclude=过期&token=xxx"

# 获取历史列表
curl "http://localhost:8080/api/configs" -H "Authorization: Bearer your-api-key"
```

## 🐳 Docker 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `API_KEY` | 访问接口所需的密钥 | - |
| `PORT` | 容器外暴露端口 | `8080` |
| `WEB_PATH` | Web 路径前缀（留空为根路径） | `""` |

> [!TIP]
> 如果你在 Hugging Face Spaces 部署，请确保 `WEB_PATH` 为空，这样你可以直接通过 Space 域名访问管理界面。

## 📄 License

MIT

## 🙏 鸣谢与致敬

本站点的核心功能与逻辑深受以下开源项目及贡献者的启发与支持，特此致谢：

- **[ACL4SSR](https://github.com/ACL4SSR/ACL4SSR)**：提供了极富盛名的规则模板库，是本项目转换逻辑的核心基石。
- **[subconverter](https://github.com/tindy2013/subconverter)**：作为同领域的先驱，为代理转换领域的通用标准提供了极具参考价值的范式。
- **[Gin Framework](https://github.com/gin-gonic/gin)**：为本项目提供高性能的路由驱动。
- **[Cobra](https://github.com/spf13/cobra)**：为命令行界面提供强力支持。

同时感谢所有开源协议提供的依赖组件，让这个项目能够快速孵化。
