---
title: Subconverter
emoji: 🔄
colorFrom: blue
colorTo: green
sdk: docker
pinned: false
---

# Subconverter - 订阅转换工具

一个功能强大的代理订阅合并和转换工具，支持多种客户端格式。

## ✨ 功能特性

- ✅ **多订阅合并** - 同时合并多个订阅源
- ✅ **12 种客户端格式** - Clash、Surge、Sing-box、Quantumult X 等
- ✅ **ACL4SSR 规则库** - 5 种在线规则模板
- ✅ **短链接配置** - 可管理的订阅配置
- ✅ **现代化 UI** - 优雅的 Web 界面

## 🚀 快速开始

### Web 界面

直接访问本 Space 的 URL，使用 Web 界面进行订阅转换。

### API 调用

```bash
# 订阅转换
curl "https://YOUR-SPACE.hf.space/api/sub?target=clash&url=订阅链接&token=API_KEY"

# 创建配置
curl -X POST https://YOUR-SPACE.hf.space/api/config \
  -H "Authorization: Bearer API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"urls":["订阅1"],"target":"clash","config":"acl4ssr_online_full"}'
```

## 🔑 API 密钥

API 密钥已配置在 Space 的 Secrets 中。如需使用，请联系 Space 所有者获取。

## 📖 文档

- [完整文档](https://github.com/Chen-ce/subconverter)
- [API 参考](https://github.com/Chen-ce/subconverter#api)
- [部署指南](https://github.com/Chen-ce/subconverter/blob/main/HUGGINGFACE_DEPLOYMENT.md)

## 🛠️ 支持的客户端

- Clash / Clash for Windows
- Sing-box
- Surge 2/3/4
- Quantumult X
- Loon
- Surfboard
- V2Ray / Shadowsocks / SSR

## 📚 规则模板

- ACL4SSR 在线基础版
- ACL4SSR 在线完整版
- ACL4SSR 在线精简版
- ACL4SSR 强化去广告版
- ACL4SSR 无自动测速版

## 🔗 相关链接

- [GitHub 仓库](https://github.com/Chen-ce/subconverter)
- [问题反馈](https://github.com/Chen-ce/subconverter/issues)

## 📄 License

MIT License
