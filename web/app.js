// 页面加载时从 localStorage 恢复配置
window.addEventListener('DOMContentLoaded', () => {
    loadSavedConfig();
    updateConfigVisibility();

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

    // 支持规则配置的格式
    const supportsConfig = ['clash', 'surge&ver=2', 'surge&ver=3', 'surge&ver=4',
        'surfboard', 'quanx', 'quan', 'loon'];

    if (supportsConfig.some(f => format.startsWith(f) || format === f)) {
        configGroup.style.display = 'block';

        // 根据格式更新提示
        if (format.includes('surge') || format === 'surfboard') {
            configHint.textContent = '适用于 Surge/Surfboard 的规则配置';
        } else if (format === 'quanx' || format === 'quan') {
            configHint.textContent = '适用于 Quantumult (X) 的规则配置';
        } else if (format === 'loon') {
            configHint.textContent = '适用于 Loon 的规则配置';
        } else {
            configHint.textContent = 'ACL4SSR 包含 Netflix、YouTube、ChatGPT 等分组';
        }
    } else {
        configGroup.style.display = 'none';
    }
}

// 转换订阅
async function convert() {
    const apiKey = document.getElementById('apiKey').value.trim();
    const subscriptions = document.getElementById('subscriptions').value.trim();
    const nodes = document.getElementById('nodes').value.trim();
    const outputFormat = document.getElementById('outputFormat').value;
    const configTemplate = document.getElementById('configTemplate').value;
    const includeFilter = document.getElementById('includeFilter').value.trim();
    const excludeFilter = document.getElementById('excludeFilter').value.trim();

    // 验证输入
    if (!apiKey) {
        showError('请输入 API 密钥');
        return;
    }

    if (!subscriptions && !nodes) {
        showError('请至少输入一个订阅链接或节点');
        return;
    }

    // 保存配置
    saveConfig();

    // 构建 API URL
    const params = new URLSearchParams();
    params.set('target', outputFormat);

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

    // 显示加载状态
    showLoading();

    try {
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

        // 生成订阅链接
        const subscriptionUrl = `${window.location.origin}/api/sub?${params.toString()}`;

        // 显示结果
        showResult(subscriptionUrl, result);

    } catch (error) {
        showError('网络错误：' + error.message);
    }
}

// 显示加载状态
function showLoading() {
    const resultSection = document.getElementById('resultSection');
    const resultInfo = document.getElementById('resultInfo');

    resultSection.style.display = 'block';
    resultInfo.innerHTML = '<div class="loading">⏳ 正在转换中...</div>';
}

// 显示结果
function showResult(url, content) {
    const resultSection = document.getElementById('resultSection');
    const resultUrl = document.getElementById('resultUrl');
    const resultInfo = document.getElementById('resultInfo');

    resultSection.style.display = 'block';
    resultUrl.value = url;

    // 统计节点数量
    const nodeCount = (content.match(/- name:/g) || []).length;

    resultInfo.innerHTML = `
        <div class="success">
            ✅ 转换成功！共 ${nodeCount} 个节点
            <br>
            <small>请复制上方链接到 Clash 或其他客户端使用</small>
        </div>
    `;

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
