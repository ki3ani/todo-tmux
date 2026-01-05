// Vault App JavaScript

const API = '/api';
let todos = [];
let vaultItems = [];
let allTags = [];

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    setupTabs();
    loadTodos();
    loadCategories();
    loadVaultItems();
    loadAllTags();
    setupEventListeners();
    maybeShowResurface();
});

// Tab Navigation
function setupTabs() {
    document.querySelectorAll('.tab').forEach(tab => {
        tab.addEventListener('click', () => {
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
            tab.classList.add('active');
            document.getElementById(tab.dataset.view + '-view').classList.add('active');
        });
    });
}

function setupEventListeners() {
    // Todo form
    document.getElementById('add-form').addEventListener('submit', handleAdd);
    document.getElementById('filter-status').addEventListener('change', loadTodos);
    document.getElementById('filter-priority').addEventListener('change', loadTodos);
    document.getElementById('filter-category').addEventListener('change', loadTodos);
    document.getElementById('filter-search').addEventListener('input', debounce(loadTodos, 300));
    document.getElementById('edit-form').addEventListener('submit', handleEdit);
    document.getElementById('edit-modal').addEventListener('click', (e) => {
        if (e.target.id === 'edit-modal') closeModal();
    });

    // Vault form
    document.getElementById('vault-add-form').addEventListener('submit', handleVaultAdd);
    document.getElementById('vault-input').addEventListener('input', debounce(handleVaultInputPreview, 500));
    document.getElementById('vault-filter-type').addEventListener('change', loadVaultItems);
    document.getElementById('vault-search').addEventListener('input', debounce(loadVaultItems, 300));
    document.getElementById('vault-edit-form').addEventListener('submit', handleVaultEdit);
    document.getElementById('vault-edit-modal').addEventListener('click', (e) => {
        if (e.target.id === 'vault-edit-modal') closeVaultModal();
    });
}

// ==================== TODOS ====================

async function loadTodos() {
    const status = document.getElementById('filter-status').value;
    const priority = document.getElementById('filter-priority').value;
    const category = document.getElementById('filter-category').value;
    const search = document.getElementById('filter-search').value;

    const params = new URLSearchParams();
    if (status) params.set('status', status);
    if (priority) params.set('priority', priority);
    if (category) params.set('category', category);
    if (search) params.set('search', search);

    const response = await fetch(`${API}/todos?${params}`);
    todos = await response.json();
    renderTodos();
}

async function loadCategories() {
    const response = await fetch(`${API}/categories`);
    const categories = await response.json();
    const select = document.getElementById('filter-category');
    select.innerHTML = '<option value="">Any Category</option>';
    categories.forEach(cat => {
        const option = document.createElement('option');
        option.value = cat;
        option.textContent = cat;
        select.appendChild(option);
    });
}

function renderTodos() {
    const list = document.getElementById('todo-list');
    const empty = document.getElementById('empty-state');

    if (todos.length === 0) {
        list.innerHTML = '';
        empty.style.display = 'block';
        return;
    }

    empty.style.display = 'none';
    list.innerHTML = todos.map(todo => renderTodoItem(todo)).join('');
}

function renderTodoItem(todo) {
    const isOverdue = todo.due_date && new Date(todo.due_date) < new Date() && !todo.done;
    return `
        <li class="todo-item ${todo.done ? 'done' : ''}">
            <button class="todo-checkbox ${todo.done ? 'checked' : ''}"
                    onclick="toggleTodo(${todo.id}, ${!todo.done})"></button>
            <div class="todo-content">
                <div class="todo-task">${escapeHtml(todo.task)}</div>
                <div class="todo-meta">
                    <span class="todo-priority ${todo.priority}">${todo.priority}</span>
                    ${todo.category ? `<span class="todo-category">#${escapeHtml(todo.category)}</span>` : ''}
                    ${todo.due_date ? `<span class="todo-due ${isOverdue ? 'overdue' : ''}">Due: ${todo.due_date}</span>` : ''}
                </div>
            </div>
            <div class="todo-actions">
                <button class="btn-edit" onclick="openEditModal(${todo.id})">✎</button>
                <button class="btn-delete" onclick="deleteTodo(${todo.id})">✕</button>
            </div>
        </li>
    `;
}

async function handleAdd(e) {
    e.preventDefault();
    const task = document.getElementById('task-input').value.trim();
    const priority = document.getElementById('priority-input').value;
    const category = document.getElementById('category-input').value.trim();
    const dueDate = document.getElementById('due-input').value;

    if (!task) return;

    await fetch(`${API}/todos`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ task, priority, category, due_date: dueDate })
    });

    document.getElementById('task-input').value = '';
    document.getElementById('category-input').value = '';
    document.getElementById('due-input').value = '';
    document.getElementById('priority-input').value = 'medium';

    loadTodos();
    loadCategories();
}

async function toggleTodo(id, done) {
    await fetch(`${API}/todos/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ done })
    });
    loadTodos();
}

async function deleteTodo(id) {
    if (!confirm('Delete this todo?')) return;
    await fetch(`${API}/todos/${id}`, { method: 'DELETE' });
    loadTodos();
    loadCategories();
}

function openEditModal(id) {
    const todo = todos.find(t => t.id === id);
    if (!todo) return;
    document.getElementById('edit-id').value = todo.id;
    document.getElementById('edit-task').value = todo.task;
    document.getElementById('edit-priority').value = todo.priority;
    document.getElementById('edit-category').value = todo.category || '';
    document.getElementById('edit-due').value = todo.due_date || '';
    document.getElementById('edit-modal').style.display = 'flex';
}

function closeModal() {
    document.getElementById('edit-modal').style.display = 'none';
}

async function handleEdit(e) {
    e.preventDefault();
    const id = document.getElementById('edit-id').value;
    const todo = todos.find(t => t.id === parseInt(id));

    await fetch(`${API}/todos/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            task: document.getElementById('edit-task').value,
            done: todo.done,
            priority: document.getElementById('edit-priority').value,
            category: document.getElementById('edit-category').value,
            due_date: document.getElementById('edit-due').value
        })
    });

    closeModal();
    loadTodos();
    loadCategories();
}

// ==================== VAULT ====================

async function loadVaultItems() {
    const type = document.getElementById('vault-filter-type').value;
    const search = document.getElementById('vault-search').value;

    const params = new URLSearchParams();
    if (type) params.set('type', type);
    if (search) params.set('search', search);

    const response = await fetch(`${API}/vault?${params}`);
    vaultItems = await response.json();
    renderVaultItems();
}

async function loadAllTags() {
    const response = await fetch(`${API}/tags`);
    allTags = await response.json();
    renderTagsFilter();
}

function renderTagsFilter() {
    const container = document.getElementById('vault-tags-filter');
    if (allTags.length === 0) {
        container.innerHTML = '';
        return;
    }
    container.innerHTML = allTags.map(tag =>
        `<span class="tag" onclick="filterByTag('${escapeHtml(tag.name)}')">#${escapeHtml(tag.name)}</span>`
    ).join('');
}

function filterByTag(tagName) {
    document.getElementById('vault-search').value = tagName;
    loadVaultItems();
}

function renderVaultItems() {
    const container = document.getElementById('vault-items');
    const empty = document.getElementById('vault-empty');

    if (vaultItems.length === 0) {
        container.innerHTML = '';
        empty.style.display = 'block';
        return;
    }

    empty.style.display = 'none';
    container.innerHTML = vaultItems.map(item => renderVaultItem(item)).join('');
}

function renderVaultItem(item) {
    const title = item.meta_title || item.title || truncate(item.content, 60);
    const typeLabel = {
        'tweet': 'Tweet', 'tiktok': 'TikTok', 'youtube': 'YouTube',
        'article': 'Article', 'note': 'Note'
    }[item.content_type] || item.content_type;

    return `
        <div class="vault-item ${item.pinned ? 'pinned' : ''}">
            <div class="vault-item-header">
                <span class="type-badge ${item.content_type}">${typeLabel}</span>
                ${item.pinned ? '<span class="pin-badge">PINNED</span>' : ''}
            </div>
            ${item.meta_thumbnail ? `<img class="vault-item-thumb" src="${item.meta_thumbnail}" alt="">` : ''}
            <div class="vault-item-title">${escapeHtml(title)}</div>
            ${item.meta_description ? `<div class="vault-item-desc">${escapeHtml(truncate(item.meta_description, 120))}</div>` : ''}
            ${item.meta_author ? `<div class="vault-item-author">${escapeHtml(item.meta_author)}</div>` : ''}
            ${item.tags && item.tags.length > 0 ? `
                <div class="vault-item-tags">
                    ${item.tags.map(t => `<span class="tag">#${escapeHtml(t.name)}</span>`).join('')}
                </div>
            ` : ''}
            <div class="vault-item-actions">
                ${item.url ? `<a href="${item.url}" target="_blank" class="btn-open">Open</a>` : ''}
                <button class="btn-pin ${item.pinned ? 'active' : ''}" onclick="toggleVaultPin(${item.id}, ${!item.pinned})">
                    ${item.pinned ? 'Unpin' : 'Pin'}
                </button>
                <button class="btn-edit" onclick="openVaultEditModal(${item.id})">Edit</button>
                <button class="btn-archive" onclick="archiveVaultItem(${item.id})">Archive</button>
            </div>
        </div>
    `;
}

async function handleVaultInputPreview() {
    const input = document.getElementById('vault-input').value.trim();
    const preview = document.getElementById('input-preview');

    if (input.length < 5) {
        preview.classList.remove('visible');
        return;
    }

    try {
        const response = await fetch(`${API}/vault/detect`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ content: input })
        });
        const data = await response.json();

        const typeLabel = {
            'tweet': 'Tweet', 'tiktok': 'TikTok', 'youtube': 'YouTube',
            'article': 'Article', 'note': 'Note'
        }[data.content_type] || data.content_type;

        let html = `<div class="preview-type"><span class="type-badge ${data.content_type}">${typeLabel}</span></div>`;
        if (data.meta_thumbnail) {
            html += `<img class="preview-thumb" src="${data.meta_thumbnail}" alt="">`;
        }
        if (data.meta_title) {
            html += `<div class="preview-title">${escapeHtml(data.meta_title)}</div>`;
        }

        preview.innerHTML = html;
        preview.classList.add('visible');
    } catch (e) {
        preview.classList.remove('visible');
    }
}

async function handleVaultAdd(e) {
    e.preventDefault();
    const content = document.getElementById('vault-input').value.trim();
    const tagsInput = document.getElementById('vault-tags').value.trim();
    const pinned = document.getElementById('vault-pin').checked;

    if (!content) return;

    const tags = tagsInput ? tagsInput.split(',').map(t => t.trim()).filter(t => t) : [];

    await fetch(`${API}/vault`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content, tags, pinned })
    });

    document.getElementById('vault-input').value = '';
    document.getElementById('vault-tags').value = '';
    document.getElementById('vault-pin').checked = false;
    document.getElementById('input-preview').classList.remove('visible');

    loadVaultItems();
    loadAllTags();
}

async function toggleVaultPin(id, pinned) {
    await fetch(`${API}/vault/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ pinned })
    });
    loadVaultItems();
}

async function archiveVaultItem(id) {
    if (!confirm('Archive this item?')) return;
    await fetch(`${API}/vault/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ archived: true })
    });
    loadVaultItems();
}

async function deleteVaultItem(id) {
    if (!confirm('Delete this item?')) return;
    await fetch(`${API}/vault/${id}`, { method: 'DELETE' });
    loadVaultItems();
    loadAllTags();
}

function openVaultEditModal(id) {
    const item = vaultItems.find(i => i.id === id);
    if (!item) return;
    document.getElementById('vault-edit-id').value = item.id;
    document.getElementById('vault-edit-title').value = item.title || item.meta_title || '';
    document.getElementById('vault-edit-content').value = item.content;
    document.getElementById('vault-edit-tags').value = item.tags ? item.tags.map(t => t.name).join(', ') : '';
    document.getElementById('vault-edit-modal').style.display = 'flex';
}

function closeVaultModal() {
    document.getElementById('vault-edit-modal').style.display = 'none';
}

async function handleVaultEdit(e) {
    e.preventDefault();
    const id = document.getElementById('vault-edit-id').value;
    const item = vaultItems.find(i => i.id === parseInt(id));
    const tagsInput = document.getElementById('vault-edit-tags').value.trim();
    const tags = tagsInput ? tagsInput.split(',').map(t => t.trim()).filter(t => t) : [];

    await fetch(`${API}/vault/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            title: document.getElementById('vault-edit-title').value,
            content: document.getElementById('vault-edit-content').value,
            pinned: item.pinned,
            archived: item.archived,
            tags
        })
    });

    closeVaultModal();
    loadVaultItems();
    loadAllTags();
}

// Resurface
async function loadResurface() {
    try {
        const response = await fetch(`${API}/vault/resurface`);
        if (!response.ok) return;
        const item = await response.json();

        const title = item.meta_title || item.title || truncate(item.content, 60);
        const content = document.getElementById('resurface-content');
        content.innerHTML = `
            <div class="vault-item-title">${escapeHtml(title)}</div>
            ${item.meta_author ? `<div class="vault-item-author">${escapeHtml(item.meta_author)}</div>` : ''}
            ${item.url ? `<a href="${item.url}" target="_blank" style="color:#e94560;font-size:12px;">Open link</a>` : ''}
        `;
        document.getElementById('resurface-banner').style.display = 'block';
    } catch (e) {
        // No items to resurface
    }
}

function dismissResurface() {
    document.getElementById('resurface-banner').style.display = 'none';
}

function maybeShowResurface() {
    if (Math.random() < 0.3) {
        loadResurface();
    }
}

// Utilities
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function truncate(s, max) {
    if (!s) return '';
    if (s.length <= max) return s;
    return s.substring(0, max - 3) + '...';
}

function debounce(func, wait) {
    let timeout;
    return function(...args) {
        clearTimeout(timeout);
        timeout = setTimeout(() => func.apply(this, args), wait);
    };
}
