# 多客户端格式支持实现计划

## 当前状态

### 已实现 ✅
- Clash
- Surge (基础实现)

### 进行中 🔄
- Quantumult X
- V2Ray/Mixed

### 待实现 📋
- Surfboard
- Loon
- SS/SSR 单独格式
- Quantumult
- 其他格式

---

## 实现优先级

### 第一阶段（本次提交）
1. ✅ Web 界面更新（客户端友好命名）
2. ✅ Surge 导出器
3. 🔄 Quantumult X 导出器
4. 🔄 Mixed/V2Ray 导出器（复用 Base64）
5. 🔄 更新 API handler 支持所有格式

### 第二阶段（下次迭代）
1. Surfboard 导出器（基于 Surge）
2. Loon 导出器
3. SS/SSR 专用导出器
4. 完善规则模板支持

### 第三阶段（后续）
1. Quantumult 导出器
2. 其他小众格式
3. 完整的配置模板系统

---

## 格式映射关系

| 客户端 | target 参数 | 实现状态 | 基于 |
|--------|------------|---------|------|
| Clash | clash | ✅ | - |
| Surge 4 | surge&ver=4 | ✅ | - |
| Surge 3 | surge&ver=3 | ✅ | Surge 4 |
| Surge 2 | surge&ver=2 | ✅ | Surge 4 |
| Quantumult X | quanx | 🔄 | - |
| Quantumult | quan | 📋 | - |
| Loon | loon | 📋 | Surge |
| Surfboard | surfboard | 📋 | Surge |
| V2Ray | v2ray | 🔄 | Base64 |
| Shadowsocks | ss | 🔄 | Base64 |
| ShadowsocksR | ssr | 🔄 | Base64 |
| Mixed | mixed | 🔄 | Base64 |

---

## 技术实现

### 导出器架构
```
Exporter Interface
├── ClashExporter (已实现)
├── SurgeExporter (已实现)
├── QuantumultXExporter (进行中)
├── Base64Exporter (已实现，用于 V2Ray/SS/SSR/Mixed)
└── 其他导出器 (待实现)
```

### API Handler 更新
```go
switch target {
case "clash":
    // Clash 导出
case "surge&ver=2", "surge&ver=3", "surge&ver=4":
    // Surge 导出
case "quanx":
    // Quantumult X 导出
case "v2ray", "ss", "ssr", "mixed":
    // Base64 导出
// ... 其他格式
}
```

---

## 测试计划

### 单元测试
- [ ] Surge 导出器测试
- [ ] Quantumult X 导出器测试
- [ ] API handler 格式路由测试

### 集成测试
- [ ] Web 界面格式选择
- [ ] API 调用各种格式
- [ ] 实际客户端导入测试

---

## 文档更新

- [ ] README 更新支持的格式列表
- [ ] API 文档更新 target 参数说明
- [ ] 添加各客户端使用示例

---

## 注意事项

1. **Surge 版本差异**
   - Surge 2/3/4 配置格式略有不同
   - 当前实现基于 Surge 4，需要适配旧版本

2. **协议支持**
   - 不是所有客户端都支持所有协议
   - 需要在导出时进行协议兼容性检查

3. **规则模板**
   - 不同客户端的规则语法不同
   - 需要为每种格式准备对应的规则模板

4. **性能考虑**
   - 大量节点转换时的性能优化
   - 模板缓存机制

---

## 下一步行动

1. 完成 Quantumult X 导出器
2. 更新 API handler 支持所有格式
3. 测试各格式导出
4. 提交代码并更新文档
