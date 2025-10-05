// 全局变量
let currentPath = '/';
let currentFiles = [];
let selectedFile = null;
let currentView = 'list';

// 初始化应用
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
    setupEventListeners();
    loadFiles('/');
    loadSpaceInfo();
});

// 初始化应用
function initializeApp() {
    // 设置拖拽上传
    setupDragAndDrop();
    
    // 设置右键菜单
    setupContextMenu();
    
    // 设置视图切换
    setupViewToggle();
}

// 设置事件监听器
function setupEventListeners() {
    // 文件输入变化
    document.getElementById('fileInput').addEventListener('change', handleFileSelect);
    
    // 搜索输入
    document.getElementById('searchInput').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            performSearch();
        }
    });
    
    // 点击模态框外部关闭
    document.addEventListener('click', function(e) {
        if (e.target.classList.contains('modal')) {
            hideAllModals();
        }
    });
    
    // ESC键关闭模态框
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape') {
            hideAllModals();
        }
    });
}

// 设置拖拽上传
function setupDragAndDrop() {
    const uploadArea = document.getElementById('uploadArea');
    
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        uploadArea.addEventListener(eventName, preventDefaults, false);
        document.body.addEventListener(eventName, preventDefaults, false);
    });
    
    ['dragenter', 'dragover'].forEach(eventName => {
        uploadArea.addEventListener(eventName, highlight, false);
    });
    
    ['dragleave', 'drop'].forEach(eventName => {
        uploadArea.addEventListener(eventName, unhighlight, false);
    });
    
    uploadArea.addEventListener('drop', handleDrop, false);
    
    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }
    
    function highlight(e) {
        uploadArea.classList.add('dragover');
    }
    
    function unhighlight(e) {
        uploadArea.classList.remove('dragover');
    }
    
    function handleDrop(e) {
        const dt = e.dataTransfer;
        const files = dt.files;
        handleFiles(files);
    }
}

// 设置右键菜单
function setupContextMenu() {
    document.addEventListener('contextmenu', function(e) {
        const fileItem = e.target.closest('.file-item');
        if (fileItem) {
            e.preventDefault();
            selectedFile = getFileFromElement(fileItem);
            showContextMenu(e.pageX, e.pageY);
        }
    });
    
    document.addEventListener('click', function() {
        hideContextMenu();
    });
}

// 设置视图切换
function setupViewToggle() {
    document.querySelectorAll('.view-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            const view = this.dataset.view;
            switchView(view);
        });
    });
}

// 加载文件列表
async function loadFiles(path) {
    try {
        showLoading();
        currentPath = path;
        
        const response = await fetch(`/api/files?path=${encodeURIComponent(path)}`);
        const result = await response.json();
        
        if (result.code === 0) {
            currentFiles = result.data.files;
            renderFiles(currentFiles);
            updateBreadcrumb(path);
        } else {
            showNotification('error', result.message);
        }
    } catch (error) {
        showNotification('error', '加载文件列表失败: ' + error.message);
    } finally {
        hideLoading();
    }
}

// 渲染文件列表
function renderFiles(files) {
    const fileList = document.getElementById('fileList');
    
    if (files.length === 0) {
        fileList.innerHTML = '<div class="empty-state">此文件夹为空</div>';
        return;
    }
    
    const html = files.map(file => `
        <div class="file-item" data-id="${file.id}" onclick="selectFile('${file.id}')" ondblclick="openFile('${file.id}')">
            <div class="file-icon ${file.isDir ? 'folder' : 'file'}">
                <i class="fas ${file.isDir ? 'fa-folder' : getFileIcon(file.name)}"></i>
            </div>
            <div class="file-info">
                <div class="file-name">${escapeHtml(file.name)}</div>
                <div class="file-meta">
                    <span>${file.isDir ? '文件夹' : formatFileSize(file.size)}</span>
                    <span>${file.modTime}</span>
                </div>
            </div>
            <div class="file-actions">
                ${!file.isDir ? `<button class="action-btn" onclick="downloadFile('${file.id}')">下载</button>` : ''}
                <button class="action-btn" onclick="renameFile('${file.id}')">重命名</button>
                <button class="action-btn" onclick="deleteFile('${file.id}')">删除</button>
            </div>
        </div>
    `).join('');
    
    fileList.innerHTML = html;
}

// 获取文件图标
function getFileIcon(filename) {
    const ext = filename.split('.').pop().toLowerCase();
    const iconMap = {
        'pdf': 'fa-file-pdf',
        'doc': 'fa-file-word',
        'docx': 'fa-file-word',
        'xls': 'fa-file-excel',
        'xlsx': 'fa-file-excel',
        'ppt': 'fa-file-powerpoint',
        'pptx': 'fa-file-powerpoint',
        'jpg': 'fa-file-image',
        'jpeg': 'fa-file-image',
        'png': 'fa-file-image',
        'gif': 'fa-file-image',
        'mp4': 'fa-file-video',
        'avi': 'fa-file-video',
        'mp3': 'fa-file-audio',
        'wav': 'fa-file-audio',
        'zip': 'fa-file-archive',
        'rar': 'fa-file-archive',
        'txt': 'fa-file-alt',
        'js': 'fa-file-code',
        'html': 'fa-file-code',
        'css': 'fa-file-code',
        'json': 'fa-file-code'
    };
    return iconMap[ext] || 'fa-file';
}

// 格式化文件大小
function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 转义HTML
function escapeHtml(text) {
    const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
    };
    return text.replace(/[&<>"']/g, function(m) { return map[m]; });
}

// 更新面包屑导航
function updateBreadcrumb(path) {
    const breadcrumb = document.getElementById('breadcrumb');
    const parts = path.split('/').filter(part => part);
    
    let html = '<span class="breadcrumb-item" onclick="navigateToPath(\'/\')"><i class="fas fa-home"></i> 根目录</span>';
    
    let currentPath = '';
    parts.forEach(part => {
        currentPath += '/' + part;
        html += ` <i class="fas fa-chevron-right"></i> <span class="breadcrumb-item" onclick="navigateToPath('${currentPath}')">${escapeHtml(part)}</span>`;
    });
    
    breadcrumb.innerHTML = html;
}

// 导航到指定路径
function navigateToPath(path) {
    loadFiles(path);
}

// 选择文件
function selectFile(id) {
    // 清除之前的选择
    document.querySelectorAll('.file-item').forEach(item => {
        item.classList.remove('selected');
    });
    
    // 选择当前文件
    const fileElement = document.querySelector(`[data-id="${id}"]`);
    if (fileElement) {
        fileElement.classList.add('selected');
        selectedFile = getFileFromElement(fileElement);
    }
}

// 打开文件/文件夹
function openFile(id) {
    const file = currentFiles.find(f => f.id === id);
    if (file && file.isDir) {
        navigateToPath(file.path);
    } else if (file) {
        downloadFile(id);
    }
}

// 从元素获取文件信息
function getFileFromElement(element) {
    const id = element.dataset.id;
    return currentFiles.find(f => f.id === id);
}

// 显示上传模态框
function showUploadModal() {
    document.getElementById('uploadModal').classList.add('show');
}

// 隐藏上传模态框
function hideUploadModal() {
    document.getElementById('uploadModal').classList.remove('show');
    resetUploadForm();
}

// 显示搜索模态框
function showSearch() {
    document.getElementById('searchModal').classList.add('show');
    document.getElementById('searchInput').focus();
}

// 隐藏搜索模态框
function hideSearchModal() {
    document.getElementById('searchModal').classList.remove('show');
    document.getElementById('searchResults').innerHTML = '';
}

// 隐藏所有模态框
function hideAllModals() {
    document.querySelectorAll('.modal').forEach(modal => {
        modal.classList.remove('show');
    });
    resetUploadForm();
}

// 重置上传表单
function resetUploadForm() {
    document.getElementById('fileInput').value = '';
    document.getElementById('uploadProgress').style.display = 'none';
    document.getElementById('uploadArea').style.display = 'block';
}

// 处理文件选择
function handleFileSelect(e) {
    const files = e.target.files;
    handleFiles(files);
}

// 处理文件
function handleFiles(files) {
    if (files.length === 0) return;
    
    // 显示进度条
    document.getElementById('uploadArea').style.display = 'none';
    document.getElementById('uploadProgress').style.display = 'block';
    
    // 上传文件
    uploadFiles(files);
}

// 上传文件
async function uploadFiles(files) {
    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        await uploadSingleFile(file, i + 1, files.length);
    }
    
    // 上传完成
    hideUploadModal();
    showNotification('success', '文件上传完成');
    loadFiles(currentPath); // 刷新文件列表
}

// 上传单个文件
async function uploadSingleFile(file, current, total) {
    try {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('path', currentPath);
        
        const response = await fetch('/api/files/upload', {
            method: 'POST',
            body: formData
        });
        
        const result = await response.json();
        
        if (result.code !== 0) {
            throw new Error(result.message);
        }
        
        // 更新进度
        const progress = (current / total) * 100;
        updateProgress(progress);
        
    } catch (error) {
        showNotification('error', `上传 ${file.name} 失败: ${error.message}`);
    }
}

// 更新进度条
function updateProgress(percent) {
    document.getElementById('progressFill').style.width = percent + '%';
    document.getElementById('progressText').textContent = Math.round(percent) + '%';
}

// 下载文件
function downloadFile(id) {
    if (!id && selectedFile) {
        id = selectedFile.id;
    }
    
    if (id) {
        window.open(`/api/files/${id}/download`, '_blank');
    }
}

// 创建文件夹
async function createFolder() {
    const name = prompt('请输入文件夹名称:');
    if (!name) return;
    
    try {
        const response = await fetch('/api/files/mkdir', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                path: currentPath,
                name: name
            })
        });
        
        const result = await response.json();
        
        if (result.code === 0) {
            showNotification('success', '文件夹创建成功');
            loadFiles(currentPath);
        } else {
            showNotification('error', result.message);
        }
    } catch (error) {
        showNotification('error', '创建文件夹失败: ' + error.message);
    }
}

// 重命名文件
async function renameFile(id) {
    if (!id && selectedFile) {
        id = selectedFile.id;
    }
    
    const file = currentFiles.find(f => f.id === id);
    if (!file) return;
    
    const newName = prompt('请输入新名称:', file.name);
    if (!newName || newName === file.name) return;
    
    try {
        const response = await fetch(`/api/files/${id}/rename`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                newName: newName
            })
        });
        
        const result = await response.json();
        
        if (result.code === 0) {
            showNotification('success', '重命名成功');
            loadFiles(currentPath);
        } else {
            showNotification('error', result.message);
        }
    } catch (error) {
        showNotification('error', '重命名失败: ' + error.message);
    }
}

// 删除文件
async function deleteFile(id) {
    if (!id && selectedFile) {
        id = selectedFile.id;
    }
    
    const file = currentFiles.find(f => f.id === id);
    if (!file) return;
    
    if (!confirm(`确定要删除 "${file.name}" 吗？`)) return;
    
    try {
        const response = await fetch(`/api/files/${id}`, {
            method: 'DELETE'
        });
        
        const result = await response.json();
        
        if (result.code === 0) {
            showNotification('success', '删除成功');
            loadFiles(currentPath);
        } else {
            showNotification('error', result.message);
        }
    } catch (error) {
        showNotification('error', '删除失败: ' + error.message);
    }
}

// 移动文件
function moveFile() {
    if (!selectedFile) return;
    
    const targetPath = prompt('请输入目标路径:');
    if (!targetPath) return;
    
    // TODO: 实现移动文件功能
    showNotification('info', '移动功能正在开发中');
}

// 执行搜索
async function performSearch() {
    const keyword = document.getElementById('searchInput').value.trim();
    if (!keyword) return;
    
    try {
        const response = await fetch(`/api/search?keyword=${encodeURIComponent(keyword)}&path=${encodeURIComponent(currentPath)}`);
        const result = await response.json();
        
        if (result.code === 0) {
            renderSearchResults(result.data.results);
        } else {
            showNotification('error', result.message);
        }
    } catch (error) {
        showNotification('error', '搜索失败: ' + error.message);
    }
}

// 渲染搜索结果
function renderSearchResults(results) {
    const container = document.getElementById('searchResults');
    
    if (results.length === 0) {
        container.innerHTML = '<div class="empty-state">未找到匹配的文件</div>';
        return;
    }
    
    const html = results.map(file => `
        <div class="file-item" onclick="navigateToPath('${file.path}')">
            <div class="file-icon ${file.isDir ? 'folder' : 'file'}">
                <i class="fas ${file.isDir ? 'fa-folder' : getFileIcon(file.name)}"></i>
            </div>
            <div class="file-info">
                <div class="file-name">${escapeHtml(file.name)}</div>
                <div class="file-meta">
                    <span>${file.path}</span>
                    <span>${file.modTime}</span>
                </div>
            </div>
        </div>
    `).join('');
    
    container.innerHTML = html;
}

// 刷新文件列表
function refreshFiles() {
    loadFiles(currentPath);
}

// 切换视图
function switchView(view) {
    currentView = view;
    
    // 更新按钮状态
    document.querySelectorAll('.view-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    document.querySelector(`[data-view="${view}"]`).classList.add('active');
    
    // 更新文件列表样式
    const fileList = document.getElementById('fileList');
    if (view === 'grid') {
        fileList.classList.add('grid-view');
    } else {
        fileList.classList.remove('grid-view');
    }
}

// 排序文件
function sortFiles() {
    const sortBy = document.getElementById('sortBy').value;
    
    const sortedFiles = [...currentFiles].sort((a, b) => {
        // 文件夹优先
        if (a.isDir && !b.isDir) return -1;
        if (!a.isDir && b.isDir) return 1;
        
        switch (sortBy) {
            case 'name':
                return a.name.localeCompare(b.name);
            case 'size':
                return b.size - a.size;
            case 'modTime':
                return new Date(b.modTime) - new Date(a.modTime);
            default:
                return 0;
        }
    });
    
    renderFiles(sortedFiles);
}

// 显示右键菜单
function showContextMenu(x, y) {
    const menu = document.getElementById('contextMenu');
    menu.style.display = 'block';
    menu.style.left = x + 'px';
    menu.style.top = y + 'px';
}

// 隐藏右键菜单
function hideContextMenu() {
    document.getElementById('contextMenu').style.display = 'none';
}

// 加载存储空间信息
async function loadSpaceInfo() {
    try {
        const response = await fetch('/api/space');
        const result = await response.json();
        
        if (result.code === 0) {
            const { total, used, free } = result.data;
            const usedPercent = (used / total) * 100;
            
            document.getElementById('spaceUsed').style.width = usedPercent + '%';
            document.getElementById('spaceText').textContent = 
                `已用 ${formatFileSize(used)} / 总计 ${formatFileSize(total)}`;
        }
    } catch (error) {
        console.error('加载存储空间信息失败:', error);
    }
}

// 显示加载状态
function showLoading() {
    document.getElementById('fileList').innerHTML = `
        <div class="loading">
            <i class="fas fa-spinner fa-spin"></i>
            加载中...
        </div>
    `;
}

// 隐藏加载状态
function hideLoading() {
    // 加载状态会被文件列表替换
}

// 显示通知
function showNotification(type, message) {
    const notifications = document.getElementById('notifications');
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    
    notifications.appendChild(notification);
    
    // 3秒后自动移除
    setTimeout(() => {
        notification.remove();
    }, 3000);
}