// 配置编辑器 - configs.js

// 页面加载
window.addEventListener('DOMContentLoaded', () => {
    loadSavedApiKey();

    // 监听输入框，支持粘贴后自动解析
    const configInput = document.getElementById('configInput');
    configInput.addEventListener('input', (e) => {
        const val = e.target.value.trim();
        // 如果是长链接且包含特征参数，自动加载
        if (val.startsWith('http') && (val.includes('url=') || val.includes('token='))) {
            loadConfig();
        }
    });

    // 监听 API 密钥输入，自动刷新列表
    const apiKeyInput = document.getElementById('apiKey');
    apiKeyInput.addEventListener('change', () => {
        saveApiKey();
        fetchConfigList();
    });

    // 检查是否从首页跳转过来
    const lastConfigId = localStorage.getItem('lastConfigId');
    if (lastConfigId) {
        document.getElementById('configInput').value = lastConfigId;
        localStorage.removeItem('lastConfigId');
        // 自动加载
        setTimeout(() => loadConfig(), 500);
    }

    // 尝试加载初始列表
    const initialKey = apiKeyInput.value.trim();
    if (initialKey) {
        setTimeout(() => fetchConfigList(), 300);
    }
});

// 加载保存的 API 密钥
function loadSavedApiKey() {
    const savedApiKey = localStorage.getItem('apiKey');
    if (savedApiKey) {
        document.getElementById('apiKey').value = savedApiKey;
    }
}

// 保存 API 密钥
function saveApiKey() {
    const apiKey = document.getElementById('apiKey').value;
    if (apiKey) {
        localStorage.setItem('apiKey', apiKey);
    }
}

// 获取配置列表
async function fetchConfigList() {
    const apiKey = document.getElementById('apiKey').value.trim();
    const configList = document.getElementById('configList');

    if (!apiKey) {
        configList.innerHTML = '<div class="empty-state"><p>请先输入 API 密钥</p><small>以加载您的历史配置</small></div>';
        return;
    }

    const btn = document.querySelector('#historySection .btn-icon');
    if (btn) btn.style.transform = 'rotate(360deg)';
    setTimeout(() => { if (btn) btn.style.transform = ''; }, 500);

    configList.innerHTML = '<div class="empty-state"><p>🚀 正在加载...</p></div>';

    try {
        const response = await fetch('/api/configs', {
            headers: {
                'Authorization': `Bearer ${apiKey}`
            }
        });

        if (response.status === 401) {
            configList.innerHTML = '<div class="empty-state"><p>❌ 身份验证失败</p><small>请检查 API 密钥是否正确</small></div>';
            return;
        }

        if (!response.ok) {
            throw new Error('无法连接到服务器');
        }

        const data = await response.json();
        const configs = data.configs || [];

        if (configs.length === 0) {
            configList.innerHTML = '<div class="empty-state"><p>📜 暂无配置</p><small>快去首页尝试生成一个吧</small></div>';
            return;
        }

        // 按时间倒序排列
        configs.sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at));

        configList.innerHTML = '';
        configs.forEach(cfg => {
            const item = document.createElement('div');
            item.className = 'config-item fade-in';
            item.onclick = () => selectFromList(cfg.id);

            const urlsCount = cfg.urls ? cfg.urls.length : 0;
            const nodesCount = cfg.nodes ? cfg.nodes.length : 0;
            const date = new Date(cfg.updated_at).toLocaleString();

            item.innerHTML = `
                <div class="config-item-header">
                    <span class="config-item-id">${cfg.id}</span>
                    <span class="config-item-target">${cfg.target}</span>
                </div>
                <div class="config-item-info">
                    ${urlsCount} 个订阅 / ${nodesCount} 个节点
                </div>
                <div class="config-item-date">${date}</div>
            `;
            configList.appendChild(item);
        });

    } catch (error) {
        configList.innerHTML = `<div class="empty-state"><p>⚠️ 加载失败</p><small>${error.message}</small></div>`;
    }
}

// 从列表选择配置
function selectFromList(id) {
    document.getElementById('configInput').value = id;
    loadConfig();

    // 高亮当前选中项
    const items = document.querySelectorAll('.config-item');
    items.forEach(item => {
        if (item.querySelector('.config-item-id').textContent === id) {
            item.classList.add('active');
        } else {
            item.classList.remove('active');
        }
    });
}

// 加载配置
async function loadConfig() {
    const input = document.getElementById('configInput').value.trim();

    if (!input) {
        notify('warning', '请输入链接或短链接 ID');
        return;
    }

    // 判断输入类型
    if (input.startsWith('http')) {
        // 完整链接
        await loadFromLongUrl(input);
    } else {
        // 短链接 ID
        await loadFromShortLink(input);
    }
}

// 从完整链接加载
async function loadFromLongUrl(url) {
    try {
        const urlObj = new URL(url);
        const params = new URLSearchParams(urlObj.search);

        // 提取参数 (确保兼容多节点参数)
        const urls = params.get('url') ? params.get('url').split('|') : [];
        const nodes = params.getAll('node');
        const target = params.get('target') || 'clash';
        const config = params.get('config') || '';
        const include = params.get('include') || '';
        const exclude = params.get('exclude') || '';
        const token = params.get('token') || '';
        const ver = params.get('ver') || '';

        // 保存并提示发现 token
        if (token) {
            const currentKey = document.getElementById('apiKey').value.trim();
            if (currentKey !== token) {
                document.getElementById('apiKey').value = token;
                saveApiKey();
                // 识别到新 token 后尝试刷新列表
                fetchConfigList();
                notify('success', '已从链接中识别 API 密钥');
            }
        }

        // 显示配置
        showConfigEditor({
            urls,
            nodes,
            target,
            config,
            include,
            exclude,
            ver
        }, 'long');

        notify('success', '配置内容已成功提取');
    } catch (error) {
        notify('error', '链接解析失败，请检查格式');
    }
}

// 从短链接加载
async function loadFromShortLink(id) {
    const apiKey = document.getElementById('apiKey').value.trim();

    if (!apiKey) {
        notify('warning', '加载短链接配置需要验证 API 密钥');
        return;
    }

    try {
        const response = await fetch(`/api/config/${id}`, {
            headers: {
                'Authorization': `Bearer ${apiKey}`
            }
        });

        if (response.status === 401) {
            notify('error', 'API 密钥验证失败');
            return;
        }

        if (!response.ok) {
            notify('error', '该配置不存在或已过期');
            return;
        }

        const config = await response.json();

        // 显示配置
        showConfigEditor(config, 'short', id);

        notify('success', '短链接配置已加载');
    } catch (error) {
        notify('error', '加载失败: ' + error.message);
    }
}

// 显示配置编辑器
function showConfigEditor(config, type, id = null) {
    const editorSection = document.getElementById('editorSection');
    const configType = document.getElementById('configType');
    const deleteBtn = document.getElementById('deleteBtn');
    const shortLinkDisplay = document.getElementById('shortLinkDisplay');

    // 显示编辑器并清空之前的状态
    editorSection.style.display = 'block';
    shortLinkDisplay.style.display = 'none';

    // 设置类型标签
    configType.textContent = type === 'short' ? '短链接配置' : '直接转换配置 (暂未保存)';

    // 保存配置 ID
    document.getElementById('configId').value = id || '';

    // 显示/隐藏删除按钮
    deleteBtn.style.display = (type === 'short' && id) ? 'inline-block' : 'none';

    // 如果已有短链接，默认展示出来
    if (type === 'short' && id) {
        const shortUrl = `${window.location.origin}/sub/${id}`;
        const shortLinkUrl = document.getElementById('shortLinkUrl');
        shortLinkUrl.value = shortUrl;
        shortLinkDisplay.style.display = 'block';
        shortLinkDisplay.querySelector('label').textContent = '短链接地址';
    }

    // 填充订阅列表
    const urlsList = document.getElementById('urlsList');
    urlsList.innerHTML = '';
    (config.urls || []).forEach(url => {
        addUrlInput(url);
    });
    if (!config.urls || config.urls.length === 0) {
        addUrlInput();
    }

    // 填充节点列表
    const nodesList = document.getElementById('nodesList');
    nodesList.innerHTML = '';
    (config.nodes || []).forEach(node => {
        addNodeInput(node);
    });

    // 填充其他字段
    const targetSelect = document.getElementById('targetFormat');
    targetSelect.value = config.target || 'clash';

    // 处理 Surge 版本匹配
    if (config.target === 'surge' && config.ver) {
        const options = Array.from(targetSelect.options);
        const match = options.find(opt => opt.value === 'surge' && opt.getAttribute('data-ver') == config.ver);
        if (match) targetSelect.selectedIndex = match.index;
    }

    document.getElementById('configTemplate').value = config.config || '';
    document.getElementById('includeFilter').value = config.include || '';
    document.getElementById('excludeFilter').value = config.exclude || '';

    // 滚动到编辑器
    editorSection.scrollIntoView({ behavior: 'smooth' });
}

// 添加和移除动画处理
function addUrlInput(value = '') {
    const urlsList = document.getElementById('urlsList');
    const div = document.createElement('div');
    div.className = 'input-with-remove fade-in';
    div.innerHTML = `
        <div class="input-wrapper">
            <span class="input-num">${urlsList.children.length + 1}</span>
            <input type="text" class="url-input" value="${value}" placeholder="粘贴订阅链接...">
        </div>
        <div class="input-actions">
            <button class="btn btn-icon-only btn-secondary" onclick="copyItemValue(this)" title="复制">
                <span class="btn-icon">📋</span>
            </button>
            <button class="btn btn-icon-only btn-danger" onclick="removeInput(this, 'urlsList')" title="删除">
                <span class="btn-icon">✕</span>
            </button>
        </div>
    `;
    urlsList.appendChild(div);
    if (!value) div.querySelector('input').focus();
}

function addNodeInput(value = '') {
    const nodesList = document.getElementById('nodesList');
    const div = document.createElement('div');
    div.className = 'input-with-remove fade-in';
    div.innerHTML = `
        <div class="input-wrapper">
            <span class="input-num">${nodesList.children.length + 1}</span>
            <input type="text" class="node-input" value="${value}" placeholder="粘贴节点链接 (vmess/ss/trojan)...">
        </div>
        <div class="input-actions">
            <button class="btn btn-icon-only btn-secondary" onclick="copyItemValue(this)" title="复制">
                <span class="btn-icon">📋</span>
            </button>
            <button class="btn btn-icon-only btn-danger" onclick="removeInput(this, 'nodesList')" title="删除">
                <span class="btn-icon">✕</span>
            </button>
        </div>
    `;
    nodesList.appendChild(div);
    if (!value) div.querySelector('input').focus();
}

function removeInput(btn, listId) {
    const item = btn.closest('.input-with-remove');
    item.classList.add('removing');
    setTimeout(() => {
        item.remove();
        // 重新对序号排序
        const list = document.getElementById(listId);
        Array.from(list.children).forEach((child, index) => {
            child.querySelector('.input-num').textContent = index + 1;
        });
        // 如果列表空了，自动加一个空的
        if (list.children.length === 0 && listId === 'urlsList') {
            addUrlInput();
        }
    }, 300);
}

// 复制单项内容
function copyItemValue(btn) {
    const input = btn.closest('.input-with-remove').querySelector('input');
    const val = input.value;
    if (!val) return;

    navigator.clipboard.writeText(val).then(() => {
        notify('success', '内容已复制到剪贴板');
    });
}

// 重置编辑器
function resetEditor() {
    showConfirm('确定要重置编辑器吗？所有未保存的改动都将丢失。', () => {
        document.getElementById('editorSection').style.display = 'none';
        document.getElementById('configId').value = '';
        document.getElementById('configInput').value = '';
        notify('success', '编辑器已重置');
    });
}

// 预览当前配置的转换结果（直接生成 API 链接）
function previewConfig() {
    const urls = Array.from(document.querySelectorAll('.url-input'))
        .map(input => input.value.trim())
        .filter(url => url.length > 0);

    const nodes = Array.from(document.querySelectorAll('.node-input'))
        .map(input => input.value.trim())
        .filter(node => node.length > 0);

    const targetSelect = document.getElementById('targetFormat');
    const target = targetSelect.value;
    const apiKey = document.getElementById('apiKey').value.trim();

    if (urls.length === 0 && nodes.length === 0) {
        notify('warning', '请至少添加一个订阅或节点');
        return;
    }

    const params = new URLSearchParams();
    params.set('target', target);

    // 处理 Surge 版本
    if (target === 'surge') {
        const selectedOption = targetSelect.options[targetSelect.selectedIndex];
        const ver = selectedOption.getAttribute('data-ver');
        if (ver) params.set('ver', ver);
    }

    if (urls.length > 0) params.set('url', urls.join('|'));
    nodes.forEach(n => params.append('node', n));

    const config = document.getElementById('configTemplate').value;
    const include = document.getElementById('includeFilter').value.trim();
    const exclude = document.getElementById('excludeFilter').value.trim();

    if (target === 'clash' && config) params.set('config', config);
    if (include) params.set('include', include);
    if (exclude) params.set('exclude', exclude);
    if (apiKey) params.set('token', apiKey);

    const fullUrl = `${window.location.origin}/api/sub?${params.toString()}`;

    // 显示预览结果区域
    const shortLinkDisplay = document.getElementById('shortLinkDisplay');
    const shortLinkUrl = document.getElementById('shortLinkUrl');

    shortLinkUrl.value = fullUrl;
    shortLinkDisplay.style.display = 'block';
    shortLinkDisplay.querySelector('label').textContent = '即时预览链接 (带参数)';

    notify('success', '已生成直接转换链接');
    shortLinkDisplay.scrollIntoView({ behavior: 'smooth' });
}

// 保存配置 (创建或更新)
async function saveConfig() {
    const configId = document.getElementById('configId').value;
    const apiKey = document.getElementById('apiKey').value.trim();

    if (!apiKey) {
        notify('warning', '保存配置必须提供 API 密钥');
        return;
    }

    const urls = Array.from(document.querySelectorAll('.url-input'))
        .map(input => input.value.trim())
        .filter(url => url.length > 0);

    const nodes = Array.from(document.querySelectorAll('.node-input'))
        .map(input => input.value.trim())
        .filter(node => node.length > 0);

    if (urls.length === 0 && nodes.length === 0) {
        notify('warning', '请至少添加一个订阅或节点');
        return;
    }

    const targetSelect = document.getElementById('targetFormat');
    const target = targetSelect.value;

    const data = {
        urls,
        nodes,
        target: target,
        include: document.getElementById('includeFilter').value.trim(),
        exclude: document.getElementById('excludeFilter').value.trim()
    };

    // 只有 Clash 才传 config
    if (target === 'clash') {
        data.config = document.getElementById('configTemplate').value;
    }

    // 处理 Surge 版本
    if (target === 'surge') {
        const selectedOption = targetSelect.options[targetSelect.selectedIndex];
        const ver = selectedOption.getAttribute('data-ver');
        if (ver) data.ver = parseInt(ver);
    }

    try {
        const url = configId ? `/api/config/${configId}` : '/api/config';
        const method = configId ? 'PUT' : 'POST';

        const response = await fetch(url, {
            method,
            headers: {
                'Authorization': `Bearer ${apiKey}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        if (response.status === 401) {
            notify('error', 'API 密钥身份验证失败');
            return;
        }

        if (!response.ok) {
            const error = await response.json();
            notify('error', error.error || '保存失败，请检查输入');
            return;
        }

        const result = await response.json();

        // 更新 UI 状态为已保存的配置
        if (!configId) {
            document.getElementById('configId').value = result.id;
            document.getElementById('configType').textContent = '短链接配置';
            document.getElementById('deleteBtn').style.display = 'inline-block';
            document.getElementById('configInput').value = result.id;
        }

        const shortUrl = `${window.location.origin}/sub/${result.id || configId}`;
        const shortLinkUrl = document.getElementById('shortLinkUrl');
        shortLinkUrl.value = shortUrl;
        document.getElementById('shortLinkDisplay').style.display = 'block';
        document.getElementById('shortLinkDisplay').querySelector('label').textContent = '短链接地址';

        // 保存成功后刷新列表
        fetchConfigList();

        notify('success', configId ? '配置已更新并同步到短链接' : '配置已保存，短链接生成成功');
    } catch (error) {
        notify('error', '网络异常: ' + error.message);
    }
}

// 删除配置
async function deleteConfig() {
    const configId = document.getElementById('configId').value;
    const apiKey = document.getElementById('apiKey').value.trim();

    if (!configId) return;

    showConfirm('确定要永久删除这个短链接配置吗？', async () => {
        try {
            const response = await fetch(`/api/config/${configId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${apiKey}`
                }
            });

            if (response.status === 401) {
                notify('error', '删除失败: API 密钥错误');
                return;
            }

            if (!response.ok) {
                notify('error', '删除失败，请稍后重试');
                return;
            }

            notify('success', '配置已成功删除');
            document.getElementById('editorSection').style.display = 'none';
            document.getElementById('configInput').value = '';

            // 删除成功后刷新列表
            fetchConfigList();
        } catch (error) {
            notify('error', '系统错误: ' + error.message);
        }
    });
}

// 复制短链接
function copyShortLink() {
    const shortLinkUrl = document.getElementById('shortLinkUrl');
    shortLinkUrl.select();
    document.execCommand('copy');
    notify('success', '链接已复制到剪贴板');
}

// 通用通知函数
function notify(type, message) {
    if (typeof window.showAlert === 'function') {
        window.showAlert(message);
        return;
    }
    if (window.Toast && typeof Toast[type] === 'function') {
        Toast[type](message);
        return;
    }
    alert(message);
}
