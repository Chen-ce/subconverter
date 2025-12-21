---
title: Subconverter
emoji: 🔄
colorFrom: blue
colorTo: green
sdk: docker
pinned: false
---

# Subconverter - 极简订阅转换工具

基于 Go 语言开发的高性能代理订阅合并与转换工具。支持多订阅合并、节点过滤、规则模板，并内置现代化的配置管理界面。

## ✨ 核心特性

- 🎯 **快速转换**：一键合并订阅与自定义节点，支持 12 种以上客户端格式。
- 🔗 **持久化短链接**：默认开启短链接生成，支持在 `/` 根路径下直接管理。
- 📱 **iOS 风格 UI**：美观现代的 Web 界面，支持移动端适配。
- 🛠️ **全能编辑器**：支持配置反向解析，支持 API Key 自动拉取历史记录。
- 🛡️ **安全可靠**：支持 API Key 认证，敏感信息不出现在 URL 参数中。

## 🚀 使用指南

### 1. 转换订阅
在本 Space 首页：
1. 输入 API Key。
2. 填入订阅链接（每行一个）。
3. 选择目标客户端（Clash, Surge, Sing-box 等）。
4. 点击“转换”，默认将生成持久化短链接。

### 2. 管理配置
点击导航栏的“配置管理”：
1. 输入 API Key 后，左侧会自动载入你所有的历史配置。
2. 点击任意配置即可载入编辑器进行二次修改或删除。

## 🔑 认证信息
本服务受 API Key 保护。如需在 API 或 Web 界面使用，请确保在 Header 或参数中携带正确的 Token。

## 📜 规则支持
- **ACL4SSR** 规则库全线支持
- 支持 **Clash** 自定义模板
- 自动适配 **Surge 2/3/4** 版本参数

## 🔗 链接
- [GitHub 源码](https://github.com/Chen-ce/subconverter)
- [完整部署文档](https://github.com/Chen-ce/subconverter/blob/main/HUGGINGFACE_DEPLOYMENT.md)
