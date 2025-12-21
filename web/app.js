// 页面加载时从 localStorage 恢复配置
window.addEventListener('DOMContentLoaded', () => {
    loadSavedConfig();
    updateConfigVisibility();
    updateTemplateDescription();

    // 监听输出格式变化
    document.getElementById('outputFormat').addEventListener('change', updateConfigVisibility);
});

// 加载保存的配置
function loadSavedConfig() {
    const savedApiKey = localStorage.getItem('apiKey');
    const savedConfig = localStorage.getItem('configTemplate');
    const savedFormat = localStorage.getItem('outputFormat');

    if (savedApiKey) {
        document.getElementById('apiKey').value = savedApiKey;
    }
    if (savedConfig) {
        document.getElementById('configTemplate').value = savedConfig;
    }
    if (savedFormat) {
        document.getElementById('outputFormat').value = savedFormat;
    }
}

// 保存配置
function saveConfig() {
    const apiKey = document.getElementById('apiKey').value;
    const config = document.getElementById('configTemplate').value;
    const format = document.getElementById('outputFormat').value;

    if (apiKey) {
        localStorage.setItem('apiKey', apiKey);
    }
    localStorage.setItem('configTemplate', config);
    localStorage.setItem('outputFormat', format);
}

// 根据输出格式显示/隐藏配置模板
function updateConfigVisibility() {
    const format = document.getElementById('outputFormat').value;
    const configGroup = document.getElementById('configGroup');
    const configHint = document.getElementById('configHint');

    const supportsConfig = format === 'clash';

    if (supportsConfig) {
        configGroup.style.display = 'block';
        if (configHint) {
            configHint.textContent = 'Clash 支持规则模板；其他格式会忽略此项';
        }
    } else {
        configGroup.style.display = 'none';
        if (configHint) {
            configHint.textContent = '选择你使用的代理客户端';
        }
    }
}

// 更新模板说明
function updateTemplateDescription() {
    const template = document.getElementById('configTemplate').value;
    const descElement = document.getElementById('templateDescription');

    const descriptions = {
        '': '使用基础配置，适合快速测试',
        'default': '✨ 简化版规则\n• 基础分流（代理/直连）\n• 适合节点较少的情况',
        'acl4ssr': '🎯 本地完整版\n• 完整的分流规则\n• Netflix、YouTube、ChatGPT 等服务分组\n• 地区节点分组（香港、日本、美国等）',
        'acl4ssr_online': '📡 在线基础版（推荐）\n• ✅ 去广告\n• ✅ 自动测速\n• ✅ 微软/苹果分流\n• 适合日常使用',
        'acl4ssr_online_full': '🚀 在线完整版（功能最全）\n• ✅ 全功能分流\n• ✅ 流媒体分组（Netflix、Disney+、YouTube等）\n• ✅ AI服务分组（ChatGPT、Bing等）\n• ✅ 游戏平台分组\n• 适合节点丰富的用户',
        'acl4ssr_online_mini': '⚡ 在线精简版\n• ✅ 基础去广告\n• ✅ 自动测速\n• ✅ 核心分流规则\n• 适合节点较少的情况',
        'acl4ssr_online_adblock': '🛡️ 强化去广告版\n• ✅✅ 增强广告拦截\n• ✅ 应用净化\n• ✅ 隐私保护\n• 适合注重去广告的用户',
        'acl4ssr_online_noauto': '🎮 无自动测速版\n• ✅ 完整分流规则\n• ❌ 无自动测速（手动选择节点）\n• 适合喜欢手动控制的用户'
    };

    if (descElement) {
        const desc = descriptions[template] || '';
        descElement.textContent = desc;
        descElement.style.display = desc ? 'block' : 'none';
    }
}

function notify(type, message) {
    if (typeof window.showAlert === 'function') {
        window.showAlert(message);
        return;
    }
    if (window.Toast && typeof Toast[type] === 'function') {
        Toast[type](message);
        return;
    }
    showError(message);
}

// 转换订阅
async function convert() {
    const apiKey = document.getElementById('apiKey').value.trim();
    const subscriptions = document.getElementById('subscriptions').value.trim();
    const nodes = document.getElementById('nodes').value.trim();
    const outputSelect = document.getElementById('outputFormat');
    const outputFormat = outputSelect.value;
    const outputVer = outputSelect.selectedOptions[0]?.dataset?.ver || '';
    const configTemplate = document.getElementById('configTemplate').value;
    const includeFilter = document.getElementById('includeFilter').value.trim();
    const excludeFilter = document.getElementById('excludeFilter').value.trim();
    const generateShortLink = document.getElementById('generateShortLink').checked;

    // 验证输入
    if (!apiKey) {
        notify('warning', '请输入 API 密钥');
        return;
    }

    if (!subscriptions && !nodes) {
        notify('warning', '请至少输入一个订阅链接或节点');
        return;
    }

    // 保存配置到 localStorage
    saveConfig();

    // 显示加载状态
    showLoading();

    try {
        if (generateShortLink) {
            // 生成短链接
            await createShortLinkConfig(apiKey, subscriptions, nodes, outputFormat, outputVer, configTemplate, includeFilter, excludeFilter);
        } else {
            // 直接转换
            await performDirectConversion(apiKey, subscriptions, nodes, outputFormat, outputVer, configTemplate, includeFilter, excludeFilter);
        }
    } catch (error) {
        showError('转换失败: ' + error.message);
    }
}

// 直接转换（不生成短链接）
async function performDirectConversion(apiKey, subscriptions, nodes, outputFormat, outputVer, configTemplate, includeFilter, excludeFilter) {
    // 构建 API URL
    const params = new URLSearchParams();
    params.set('target', outputFormat);
    if (outputFormat === 'surge' && outputVer) {
        params.set('ver', outputVer);
    }

    // 处理订阅链接
    if (subscriptions) {
        const subList = subscriptions.split('\n')
            .map(s => s.trim())
            .filter(s => s.length > 0);
        params.set('url', subList.join('|'));
    }

    // 处理单独节点
    if (nodes) {
        const nodeList = nodes.split('\n')
            .map(n => n.trim())
            .filter(n => n.length > 0);
        nodeList.forEach(node => {
            params.append('node', node);
        });
    }

    // 添加配置模板（仅 Clash 格式）
    if (outputFormat === 'clash') {
        params.set('config', configTemplate);
    }

    // 添加过滤条件
    if (includeFilter) {
        params.set('include', includeFilter);
    }
    if (excludeFilter) {
        params.set('exclude', excludeFilter);
    }

    // 添加 API 密钥
    params.set('token', apiKey);

    const response = await fetch(`/api/sub?${params.toString()}`);

    if (response.status === 401) {
        showError('API 密钥错误，请检查后重试');
        return;
    }

    if (!response.ok) {
        const error = await response.json();
        showError(error.error || '转换失败，请检查输入');
        return;
    }

    // 获取结果
    const result = await response.text();

    // 生成订阅链接 (需包含 token)
    const finalParams = new URLSearchParams(params.toString());
    if (!finalParams.has('token') && apiKey) {
        finalParams.set('token', apiKey);
    }
    const subscriptionUrl = `${window.location.origin}/api/sub?${finalParams.toString()}`;

    // 显示结果
    showResult(subscriptionUrl, result, false);
}

// 创建短链接配置
async function createShortLinkConfig(apiKey, subscriptions, nodes, outputFormat, outputVer, configTemplate, includeFilter, excludeFilter) {
    const data = {
        urls: subscriptions ? subscriptions.split('\n').map(s => s.trim()).filter(s => s.length > 0) : [],
        nodes: nodes ? nodes.split('\n').map(n => n.trim()).filter(n => n.length > 0) : [],
        target: outputFormat,
        include: includeFilter || '',
        exclude: excludeFilter || ''
    };

    // 只有 Clash 才发送配置模板
    if (outputFormat === 'clash') {
        data.config = configTemplate || '';
    }

    // 只有 Surge 才发送版本
    if (outputFormat === 'surge' && outputVer) {
        data.ver = parseInt(outputVer);
    }

    const response = await fetch('/api/config', {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${apiKey}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    });

    if (response.status === 401) {
        showError('API 密钥错误，请检查后重试');
        return;
    }

    if (!response.ok) {
        const error = await response.json();
        showError(error.error || '创建短链接失败');
        return;
    }

    const result = await response.json();
    const id = result.id;
    const currentApiKey = document.getElementById('apiKey').value.trim();
    const shortUrl = `${window.location.origin}/sub/${id}${currentApiKey ? `?token=${encodeURIComponent(currentApiKey)}` : ''}`;

    // 显示短链接结果
    showResult(shortUrl, null, true, id);
}

// 显示加载状态
function showLoading() {
    const resultSection = document.getElementById('resultSection');
    const resultInfo = document.getElementById('resultInfo');

    resultSection.style.display = 'block';
    resultInfo.innerHTML = '<div class="loading">⏳ 正在转换中...</div>';
}

// 显示结果
function showResult(url, content, isShortLink = false, configId = null) {
    const resultSection = document.getElementById('resultSection');
    const resultUrl = document.getElementById('resultUrl');
    const resultInfo = document.getElementById('resultInfo');
    const resultTag = document.getElementById('resultTag');
    const shortLinkActions = document.getElementById('shortLinkActions');

    resultSection.style.display = 'block';
    resultUrl.value = url;

    if (isShortLink) {
        // 短链接模式
        resultTag.textContent = '短链接';
        resultInfo.innerHTML = `
            <div class="success">
                ✅ 短链接创建成功！
                <br>
                <small>此链接可在配置页面管理和修改</small>
            </div>
        `;
        shortLinkActions.style.display = 'block';

        // 保存配置 ID 到 localStorage，方便跳转到配置页面
        localStorage.setItem('lastConfigId', configId);
    } else {
        // 直接转换模式
        resultTag.textContent = '订阅链接';

        // 统计节点数量
        const nodeCount = content ? (content.match(/- name:/g) || []).length : 0;
        const countText = nodeCount > 0 ? `✅ 转换成功！共 ${nodeCount} 个节点` : '✅ 转换成功！';

        resultInfo.innerHTML = `
            <div class="success">
                ${countText}
                <br>
                <small>请复制上方链接到对应客户端使用</small>
            </div>
        `;
        shortLinkActions.style.display = 'none';
    }

    // 滚动到结果
    resultSection.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
}

// 显示错误
function showError(message) {
    const resultSection = document.getElementById('resultSection');
    const resultInfo = document.getElementById('resultInfo');

    resultSection.style.display = 'block';
    resultInfo.innerHTML = `<div class="error">❌ ${message}</div>`;

    // 滚动到错误信息
    resultSection.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
}

// 复制结果
function copyResult() {
    const resultUrl = document.getElementById('resultUrl');
    resultUrl.select();
    document.execCommand('copy');

    // 显示复制成功提示
    const btn = event.target.closest('.btn-copy');
    const originalText = btn.innerHTML;
    btn.innerHTML = '<span class="btn-icon">✅</span> 已复制';
    btn.classList.add('copied');

    setTimeout(() => {
        btn.innerHTML = originalText;
        btn.classList.remove('copied');
    }, 2000);
}

// 清空表单
function clearForm() {
    if (confirm('确定要清空所有输入吗？')) {
        document.getElementById('subscriptions').value = '';
        document.getElementById('nodes').value = '';
        document.getElementById('includeFilter').value = '';
        document.getElementById('excludeFilter').value = '';
        document.getElementById('resultSection').style.display = 'none';
    }
}

// 预览转换后的原始文本内容
async function viewRawContent() {
    const url = document.getElementById('resultUrl').value;
    if (!url) return;

    showLoading();
    try {
        const response = await fetch(url);
        if (!response.ok) throw new Error('无法获取转换内容');
        const text = await response.text();

        // 使用 showAlert 展示内容（或者你可以后续添加更精美的弹窗）
        // 这里我们简单打印，或者你可以考虑添加一个 Modal
        const pre = document.createElement('pre');
        pre.style.textAlign = 'left';
        pre.style.maxHeight = '400px';
        pre.style.overflow = 'auto';
        pre.style.fontSize = '12px';
        pre.style.padding = '1rem';
        pre.style.background = 'rgba(0,0,0,0.05)';
        pre.style.borderRadius = '8px';
        pre.textContent = text;

        showConfirm('转换内容预览', () => { }, '关闭', pre.outerHTML);
    } catch (error) {
        showError('内容预览失败: ' + error.message);
    }
}
