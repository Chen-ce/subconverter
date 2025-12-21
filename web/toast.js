// Toast 通知系统
const Toast = {
    container: null,

    init() {
        if (!this.container) {
            this.container = document.createElement('div');
            this.container.id = 'toast-container';
            this.container.className = 'toast-container';
            document.body.appendChild(this.container);
        }
    },

    show(message, type = 'info', duration = 3000) {
        this.init();

        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;

        const icon = this.getIcon(type);
        toast.innerHTML = `
            <span class="toast-icon">${icon}</span>
            <span class="toast-message">${message}</span>
        `;

        this.container.appendChild(toast);

        // 触发动画
        setTimeout(() => toast.classList.add('toast-show'), 10);

        // 自动移除
        setTimeout(() => {
            toast.classList.remove('toast-show');
            setTimeout(() => toast.remove(), 300);
        }, duration);
    },

    getIcon(type) {
        const icons = {
            success: '✓',
            error: '✕',
            warning: '⚠',
            info: 'ℹ'
        };
        return icons[type] || icons.info;
    },

    success(message, duration) {
        this.show(message, 'success', duration);
    },

    error(message, duration) {
        this.show(message, 'error', duration);
    },

    warning(message, duration) {
        this.show(message, 'warning', duration);
    },

    info(message, duration) {
        this.show(message, 'info', duration);
    }
};

// 确认对话框
function showConfirm(message, onConfirm, onCancel) {
    const overlay = document.createElement('div');
    overlay.className = 'confirm-overlay';

    const dialog = document.createElement('div');
    dialog.className = 'confirm-dialog';
    dialog.innerHTML = `
        <div class="confirm-content">
            <div class="confirm-icon">⚠️</div>
            <div class="confirm-message">${message}</div>
        </div>
        <div class="confirm-actions">
            <button class="btn btn-secondary confirm-cancel">取消</button>
            <button class="btn btn-danger confirm-ok">确定</button>
        </div>
    `;

    overlay.appendChild(dialog);
    document.body.appendChild(overlay);

    // 动画
    setTimeout(() => overlay.classList.add('confirm-show'), 10);

    const close = (confirmed) => {
        overlay.classList.remove('confirm-show');
        setTimeout(() => overlay.remove(), 200);
        if (confirmed && onConfirm) onConfirm();
        if (!confirmed && onCancel) onCancel();
    };

    dialog.querySelector('.confirm-ok').onclick = () => close(true);
    dialog.querySelector('.confirm-cancel').onclick = () => close(false);
    overlay.onclick = (e) => {
        if (e.target === overlay) close(false);
    };
}

// 提示对话框
function showAlert(message, onClose) {
    const overlay = document.createElement('div');
    overlay.className = 'alert-overlay';

    const dialog = document.createElement('div');
    dialog.className = 'alert-dialog';
    dialog.innerHTML = `
        <div class="alert-content">
            <div class="alert-icon">ℹ️</div>
            <div class="alert-message">${message}</div>
        </div>
        <div class="alert-actions">
            <button class="btn btn-primary alert-ok">知道了</button>
        </div>
    `;

    overlay.appendChild(dialog);
    document.body.appendChild(overlay);

    setTimeout(() => overlay.classList.add('alert-show'), 10);

    const close = () => {
        overlay.classList.remove('alert-show');
        setTimeout(() => overlay.remove(), 200);
        if (onClose) onClose();
    };

    dialog.querySelector('.alert-ok').onclick = close;
    overlay.onclick = (e) => {
        if (e.target === overlay) close();
    };
}
