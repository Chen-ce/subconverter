# ACL4SSR 远程模板配置

本项目支持使用 ACL4SSR 的远程模板，通过 jsdelivr CDN 加速访问。

## 可用的 ACL4SSR 模板

### 基础版本
- **ACL4SSR_Online**: 在线基础版
  - URL: `https://cdn.jsdelivr.net/gh/ACL4SSR/ACL4SSR@master/Clash/config/ACL4SSR_Online.ini`
  - 特点: 去广告、自动测速、微软/苹果分流

### 完整版本
- **ACL4SSR_Online_Full**: 在线完整版
  - URL: `https://cdn.jsdelivr.net/gh/ACL4SSR/ACL4SSR@master/Clash/config/ACL4SSR_Online_Full.ini`
  - 特点: 全功能，包含所有分流规则

### 其他版本
- **ACL4SSR_Online_AdblockPlus**: 强化去广告版
- **ACL4SSR_Online_NoAuto**: 无自动测速版
- **ACL4SSR_Online_NoReject**: 无广告拦截版
- **ACL4SSR_Online_Mini**: 精简版
- **ACL4SSR_Online_Mini_AdblockPlus**: 精简强化去广告版
- **ACL4SSR_Online_Mini_NoAuto**: 精简无自动测速版
- **ACL4SSR_Online_Mini_Fallback**: 精简故障转移版
- **ACL4SSR_Online_Mini_MultiMode**: 精简多模式版

## 使用方式

### 方式 1: 在 Web 界面选择

在首页或配置编辑器中创建配置时，在“配置模板”下拉菜单选择 ACL4SSR 相关选项即可。系统会自动根据你选择的模板生成带参数的转换链接。

### 方式 2: API 调用

```bash
# 使用 ACL4SSR 在线完整版
curl "http://localhost:8080/sub?target=clash&url=订阅链接&config=acl4ssr_online_full&token=xxx"

# 使用 ACL4SSR 在线基础版
curl "http://localhost:8080/sub?target=clash&url=订阅链接&config=acl4ssr_online&token=xxx"
```

### 方式 3: 直接使用远程 URL

```bash
curl "http://localhost:8080/sub?target=clash&url=订阅链接&remote_config=https://cdn.jsdelivr.net/gh/ACL4SSR/ACL4SSR@master/Clash/config/ACL4SSR_Online_Full.ini&token=xxx"
```

## 模板特点对比

| 模板 | 去广告 | 自动测速 | 流媒体分流 | 节点数量要求 |
|------|--------|----------|-----------|-------------|
| Online | ✅ | ✅ | ✅ | 适中 |
| Online_Full | ✅ | ✅ | ✅✅ | 较多 |
| Online_Mini | ✅ | ✅ | 基础 | 较少 |
| Online_NoAuto | ✅ | ❌ | ✅ | 适中 |
| Online_AdblockPlus | ✅✅ | ✅ | ✅ | 适中 |

## CDN 加速

所有模板都通过 jsdelivr CDN 提供，确保：
- ✅ 国内访问稳定
- ✅ 自动更新规则
- ✅ 高可用性

## 自定义模板

如果需要自定义模板，可以：
1. Fork ACL4SSR 仓库
2. 修改配置文件
3. 使用你的 GitHub 仓库 URL：
   ```
   https://cdn.jsdelivr.net/gh/你的用户名/ACL4SSR@master/Clash/config/你的配置.ini
   ```
