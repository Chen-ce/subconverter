// 页面加载时初始化
window.addEventListener('DOMContentLoaded', () => {
    loadSavedApiKey();
    loadConfigs();
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

// 获取 API 密钥
function getApiKey() {
    const apiKey = document.getElementById('apiKey').value.trim();
    if (!apiKey) {
        alert('请先输入 API 密钥');
        return null;
    }
    saveApiKey();
    return apiKey;
}

// 加载配置列表
async function loadConfigs() {
    const apiKey = getApiKey();
    if (!apiKey) return;

    try {
        const response = await fetch('/api/configs', {
            headers: {
                'Authorization': `Bearer ${apiKey}`
            }
        });

        if (response.status === 401) {
            alert('API 密钥错误');
            return;
        }

        if (!response.ok) {
            throw new Error('加载配置失败');
        }

        const data = await response.json();
        displayConfigs(data.configs || []);

    } catch (error) {
        alert('加载配置失败: ' + error.message);
    }
}

// 显示配置列表
function displayConfigs(configs) {
    const container = document.getElementById('configsList');

    if (configs.length === 0) {
        container.innerHTML = '<p style="text-align: center; color: #718096;">暂无配置，点击上方按钮创建</p>';
        return;
    }

    let html = '<div class="configs-grid">';

    configs.forEach(cfg => {
        const shortUrl = `${window.location.origin}/sub/${cfg.id}`;
        const createdAt = new Date(cfg.created_at).toLocaleString('zh-CN');
        const updatedAt = new Date(cfg.updated_at).toLocaleString('zh-CN');

        html += `
            <div class="config-card">
                <div class="config-header">
                    <h3>${cfg.target.toUpperCase()}</h3>
                    <span class="config-id">${cfg.id}</span>
                </div>
                <div class="config-body">
                    <p><strong>订阅数量:</strong> ${cfg.urls.length}</p>
                    <p><strong>节点数量:</strong> ${cfg.nodes ? cfg.nodes.length : 0}</p>
                    <p><strong>规则模板:</strong> ${cfg.config || '默认'}</p>
                    <p><strong>创建时间:</strong> ${createdAt}</p>
                    <p><strong>更新时间:</strong> ${updatedAt}</p>
                    <div class="config-url">
                        <input type="text" value="${shortUrl}" readonly onclick="this.select()">
                        <button class="btn-icon-only" onclick="copyUrl('${shortUrl}')" title="复制">📋</button>
                    </div>
                </div>
                <div class="config-actions">
                    <button class="btn btn-small btn-secondary" onclick="editConfig('${cfg.id}')">编辑</button>
                    <button class="btn btn-small btn-danger" onclick="deleteConfig('${cfg.id}')">删除</button>
                </div>
            </div>
        `;
    });

    html += '</div>';
    container.innerHTML = html;
}

// 显示创建模态框
function showCreateModal() {
    document.getElementById('modalTitle').textContent = '创建配置';
    document.getElementById('editingId').value = '';
    document.getElementById('modalUrls').value = '';
    document.getElementById('modalNodes').value = '';
    document.getElementById('modalTarget').value = 'clash';
    document.getElementById('modalConfig').value = '';
    document.getElementById('modalInclude').value = '';
    document.getElementById('modalExclude').value = '';
    document.getElementById('configModal').style.display = 'flex';
}

// 编辑配置
async function editConfig(id) {
    const apiKey = getApiKey();
    if (!apiKey) return;

    try {
        const response = await fetch(`/api/config/${id}`, {
            headers: {
                'Authorization': `Bearer ${apiKey}`
            }
        });

        if (!response.ok) {
            throw new Error('加载配置失败');
        }

        const cfg = await response.json();

        document.getElementById('modalTitle').textContent = '编辑配置';
        document.getElementById('editingId').value = cfg.id;
        document.getElementById('modalUrls').value = cfg.urls.join('\n');
        document.getElementById('modalNodes').value = cfg.nodes ? cfg.nodes.join('\n') : '';
        document.getElementById('modalTarget').value = cfg.target;
        document.getElementById('modalConfig').value = cfg.config || '';
        document.getElementById('modalInclude').value = cfg.include || '';
        document.getElementById('modalExclude').value = cfg.exclude || '';
        document.getElementById('configModal').style.display = 'flex';

    } catch (error) {
        alert('加载配置失败: ' + error.message);
    }
}

// 保存配置
async function saveConfig() {
    const apiKey = getApiKey();
    if (!apiKey) return;

    const id = document.getElementById('editingId').value;
    const urls = document.getElementById('modalUrls').value.split('\n')
        .map(u => u.trim())
        .filter(u => u.length > 0);
    const nodes = document.getElementById('modalNodes').value.split('\n')
        .map(n => n.trim())
        .filter(n => n.length > 0);
    const target = document.getElementById('modalTarget').value;
    const config = document.getElementById('modalConfig').value;
    const include = document.getElementById('modalInclude').value.trim();
    const exclude = document.getElementById('modalExclude').value.trim();

    if (urls.length === 0) {
        alert('请至少输入一个订阅链接');
        return;
    }

    const data = {
        urls,
        nodes,
        target,
        config,
        include,
        exclude
    };

    try {
        const url = id ? `/api/config/${id}` : '/api/config';
        const method = id ? 'PUT' : 'POST';

        const response = await fetch(url, {
            method,
            headers: {
                'Authorization': `Bearer ${apiKey}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            throw new Error('保存配置失败');
        }

        closeModal();
        loadConfigs();
        alert(id ? '配置已更新' : '配置已创建');

    } catch (error) {
        alert('保存配置失败: ' + error.message);
    }
}

// 删除配置
async function deleteConfig(id) {
    if (!confirm('确定要删除这个配置吗？')) {
        return;
    }

    const apiKey = getApiKey();
    if (!apiKey) return;

    try {
        const response = await fetch(`/api/config/${id}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${apiKey}`
            }
        });

        if (!response.ok) {
            throw new Error('删除配置失败');
        }

        loadConfigs();
        alert('配置已删除');

    } catch (error) {
        alert('删除配置失败: ' + error.message);
    }
}

// 关闭模态框
function closeModal() {
    document.getElementById('configModal').style.display = 'none';
}

// 复制 URL
function copyUrl(url) {
    navigator.clipboard.writeText(url).then(() => {
        alert('已复制到剪贴板');
    }).catch(() => {
        // 降级方案
        const input = document.createElement('input');
        input.value = url;
        document.body.appendChild(input);
        input.select();
        document.execCommand('copy');
        document.body.removeChild(input);
        alert('已复制到剪贴板');
    });
}

// 点击模态框外部关闭
window.onclick = function (event) {
    const modal = document.getElementById('configModal');
    if (event.target === modal) {
        closeModal();
    }
}
