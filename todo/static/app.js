// Todo App JavaScript

const API = '/api/todos';
let todos = [];

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    loadTodos();
    loadCategories();
    setupEventListeners();
});

function setupEventListeners() {
    // Add form
    document.getElementById('add-form').addEventListener('submit', handleAdd);

    // Filters
    document.getElementById('filter-status').addEventListener('change', loadTodos);
    document.getElementById('filter-priority').addEventListener('change', loadTodos);
    document.getElementById('filter-category').addEventListener('change', loadTodos);
    document.getElementById('filter-search').addEventListener('input', debounce(loadTodos, 300));

    // Edit form
    document.getElementById('edit-form').addEventListener('submit', handleEdit);

    // Modal close on background click
    document.getElementById('edit-modal').addEventListener('click', (e) => {
        if (e.target.id === 'edit-modal') closeModal();
    });
}

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

    const response = await fetch(`${API}?${params}`);
    todos = await response.json();
    renderTodos();
}

async function loadCategories() {
    const response = await fetch('/api/categories');
    const categories = await response.json();
    const select = document.getElementById('filter-category');

    // Keep first option
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

    await fetch(API, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ task, priority, category, due_date: dueDate })
    });

    // Clear form
    document.getElementById('task-input').value = '';
    document.getElementById('category-input').value = '';
    document.getElementById('due-input').value = '';
    document.getElementById('priority-input').value = 'medium';

    loadTodos();
    loadCategories();
}

async function toggleTodo(id, done) {
    await fetch(`${API}/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ done })
    });
    loadTodos();
}

async function deleteTodo(id) {
    if (!confirm('Delete this todo?')) return;

    await fetch(`${API}/${id}`, { method: 'DELETE' });
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

    await fetch(`${API}/${id}`, {
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

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function debounce(func, wait) {
    let timeout;
    return function(...args) {
        clearTimeout(timeout);
        timeout = setTimeout(() => func.apply(this, args), wait);
    };
}
