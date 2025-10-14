// Configuration
const API_BASE_URL = window.location.origin;
// const WS_URL = `ws://${window.location.host}/ws`;
const WS_URL = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;
const BOARD_ID = '660e8400-e29b-41d4-a716-446655440001'; // Default board ID

// State
let ws = null;
let reconnectInterval = null;
let tasks = [];
let currentEditingTask = null;

// DOM Elements
const elements = {
    connectionStatus: document.getElementById('connectionStatus'),
    activeUsers: document.getElementById('activeUsers'),
    addTaskBtn: document.getElementById('addTaskBtn'),
    taskModal: document.getElementById('taskModal'),
    closeModal: document.getElementById('closeModal'),
    cancelBtn: document.getElementById('cancelBtn'),
    taskForm: document.getElementById('taskForm'),
    modalTitle: document.getElementById('modalTitle'),
    submitBtn: document.getElementById('submitBtn'),
    toast: document.getElementById('toast'),
    toastMessage: document.getElementById('toastMessage'),
    todoList: document.getElementById('todoList'),
    inProgressList: document.getElementById('inProgressList'),
    doneList: document.getElementById('doneList'),
    todoCount: document.getElementById('todoCount'),
    inProgressCount: document.getElementById('inProgressCount'),
    doneCount: document.getElementById('doneCount')
};

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    initWebSocket();
    loadTasks();
    attachEventListeners();
});

// WebSocket Functions
function initWebSocket() {
    try {
        ws = new WebSocket(`${WS_URL}?board_id=${BOARD_ID}`);
        
        ws.onopen = () => {
            console.log('✅ WebSocket connected');
            updateConnectionStatus(true);
            clearReconnectInterval();
        };
        
        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                handleWebSocketEvent(data);
            } catch (error) {
                console.error('Failed to parse WebSocket message:', error);
            }
        };
        
        ws.onerror = (error) => {
            console.error('❌ WebSocket error:', error);
        };
        
        ws.onclose = () => {
            console.log('🔌 WebSocket disconnected');
            updateConnectionStatus(false);
            scheduleReconnect();
        };
    } catch (error) {
        console.error('Failed to initialize WebSocket:', error);
        scheduleReconnect();
    }
}

function updateConnectionStatus(connected) {
    if (connected) {
        elements.connectionStatus.classList.add('connected');
        elements.connectionStatus.classList.remove('disconnected');
        elements.connectionStatus.querySelector('span').textContent = 'Connected';
    } else {
        elements.connectionStatus.classList.remove('connected');
        elements.connectionStatus.classList.add('disconnected');
        elements.connectionStatus.querySelector('span').textContent = 'Disconnected';
    }
}

function scheduleReconnect() {
    if (!reconnectInterval) {
        reconnectInterval = setInterval(() => {
            console.log('🔄 Attempting to reconnect...');
            initWebSocket();
        }, 5000);
    }
}

function clearReconnectInterval() {
    if (reconnectInterval) {
        clearInterval(reconnectInterval);
        reconnectInterval = null;
    }
}

function handleWebSocketEvent(event) {
    console.log('📨 WebSocket event:', event);
    
    switch (event.type) {
        case 'task_created':
            if (event.task) {
                addTaskToBoard(event.task, true);
                showToast('New task created');
            }
            break;
            
        case 'task_updated':
            if (event.task) {
                updateTaskOnBoard(event.task);
                showToast('Task updated');
            }
            break;
            
        case 'task_deleted':
            if (event.task_id) {
                removeTaskFromBoard(event.task_id);
                showToast('Task deleted');
            }
            break;
    }
}

// API Functions
async function loadTasks() {
    try {
        const response = await fetch(`${API_BASE_URL}/api/boards/${BOARD_ID}/tasks`);
        if (!response.ok) throw new Error('Failed to load tasks');
        
        tasks = await response.json();
        renderAllTasks();
    } catch (error) {
        console.error('Failed to load tasks:', error);
        showToast('Failed to load tasks', 'error');
    }
}

async function createTask(taskData) {
    try {
        const response = await fetch(`${API_BASE_URL}/api/tasks`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ...taskData, board_id: BOARD_ID })
        });
        
        if (!response.ok) throw new Error('Failed to create task');
        
        return await response.json();
    } catch (error) {
        console.error('Failed to create task:', error);
        throw error;
    }
}

async function updateTask(taskId, taskData) {
    try {
        const response = await fetch(`${API_BASE_URL}/api/tasks/${taskId}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(taskData)
        });
        
        if (!response.ok) throw new Error('Failed to update task');
        
        return await response.json();
    } catch (error) {
        console.error('Failed to update task:', error);
        throw error;
    }
}

async function deleteTask(taskId) {
    try {
        const response = await fetch(`${API_BASE_URL}/api/tasks/${taskId}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) throw new Error('Failed to delete task');
    } catch (error) {
        console.error('Failed to delete task:', error);
        throw error;
    }
}

// UI Functions
function renderAllTasks() {
    elements.todoList.innerHTML = '';
    elements.inProgressList.innerHTML = '';
    elements.doneList.innerHTML = '';
    
    tasks.forEach(task => addTaskToBoard(task, false));
    updateTaskCounts();
}

function addTaskToBoard(task, isNew = false) {
    // Check if task already exists
    const existingIndex = tasks.findIndex(t => t.id === task.id);
    if (existingIndex === -1) {
        tasks.push(task);
    } else {
        tasks[existingIndex] = task;
    }
    
    const taskCard = createTaskCard(task, isNew);
    const targetList = getListByStatus(task.status);
    targetList.appendChild(taskCard);
    
    updateTaskCounts();
}

function updateTaskOnBoard(task) {
    // Remove old task card
    const oldCard = document.querySelector(`[data-task-id="${task.id}"]`);
    if (oldCard) oldCard.remove();
    
    // Update task in array
    const index = tasks.findIndex(t => t.id === task.id);
    if (index !== -1) {
        tasks[index] = task;
    }
    
    // Add updated task
    addTaskToBoard(task, false);
}

function removeTaskFromBoard(taskId) {
    const taskCard = document.querySelector(`[data-task-id="${taskId}"]`);
    if (taskCard) {
        taskCard.style.animation = 'slideOut 0.3s ease-out';
        setTimeout(() => taskCard.remove(), 300);
    }
    
    tasks = tasks.filter(t => t.id !== taskId);
    updateTaskCounts();
}

function createTaskCard(task, isNew = false) {
    const card = document.createElement('div');
    card.className = `task-card priority-${task.priority}${isNew ? ' new-task' : ''}`;
    card.dataset.taskId = task.id;
    
    const dueDate = task.due_date ? new Date(task.due_date).toLocaleDateString() : null;
    
    card.innerHTML = `
        <div class="task-header">
            <h4 class="task-title">${escapeHtml(task.title)}</h4>
            <div class="task-actions">
                <button class="task-action-btn edit-task" onclick="editTask('${task.id}')">
                    <i class="fas fa-edit"></i>
                </button>
                <button class="task-action-btn delete-task" onclick="confirmDeleteTask('${task.id}')">
                    <i class="fas fa-trash"></i>
                </button>
            </div>
        </div>
        ${task.description ? `<p class="task-description">${escapeHtml(task.description)}</p>` : ''}
        <div class="task-meta">
            <span class="task-priority">
                <span class="priority-badge ${task.priority}">${task.priority}</span>
            </span>
            ${dueDate ? `<span class="task-due-date"><i class="fas fa-calendar"></i> ${dueDate}</span>` : ''}
        </div>
    `;
    
    return card;
}

function getListByStatus(status) {
    switch (status) {
        case 'todo':
            return elements.todoList;
        case 'in_progress':
            return elements.inProgressList;
        case 'done':
            return elements.doneList;
        default:
            return elements.todoList;
    }
}

function updateTaskCounts() {
    const counts = {
        todo: tasks.filter(t => t.status === 'todo').length,
        in_progress: tasks.filter(t => t.status === 'in_progress').length,
        done: tasks.filter(t => t.status === 'done').length
    };
    
    elements.todoCount.textContent = counts.todo;
    elements.inProgressCount.textContent = counts.in_progress;
    elements.doneCount.textContent = counts.done;
    
    // Show empty states
    showEmptyState(elements.todoList, counts.todo);
    showEmptyState(elements.inProgressList, counts.in_progress);
    showEmptyState(elements.doneList, counts.done);
}

function showEmptyState(list, count) {
    const emptyState = list.querySelector('.empty-state');
    
    if (count === 0 && !emptyState) {
        const empty = document.createElement('div');
        empty.className = 'empty-state';
        empty.innerHTML = `
            <i class="fas fa-inbox"></i>
            <p>No tasks here</p>
        `;
        list.appendChild(empty);
    } else if (count > 0 && emptyState) {
        emptyState.remove();
    }
}

// Modal Functions
function openModal(task = null) {
    currentEditingTask = task;
    
    if (task) {
        elements.modalTitle.textContent = 'Edit Task';
        elements.submitBtn.innerHTML = '<i class="fas fa-save"></i> Update Task';
        populateForm(task);
    } else {
        elements.modalTitle.textContent = 'Add New Task';
        elements.submitBtn.innerHTML = '<i class="fas fa-save"></i> Save Task';
        elements.taskForm.reset();
    }
    
    elements.taskModal.classList.add('active');
}

function closeModal() {
    elements.taskModal.classList.remove('active');
    elements.taskForm.reset();
    currentEditingTask = null;
}

function populateForm(task) {
    document.getElementById('taskTitle').value = task.title;
    document.getElementById('taskDescription').value = task.description || '';
    document.getElementById('taskStatus').value = task.status;
    document.getElementById('taskPriority').value = task.priority;
    
    if (task.due_date) {
        const date = new Date(task.due_date);
        const localDate = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
        document.getElementById('taskDueDate').value = localDate.toISOString().slice(0, 16);
    }
}

async function handleFormSubmit(e) {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    const taskData = {
        title: formData.get('title'),
        description: formData.get('description') || '',
        status: formData.get('status'),
        priority: formData.get('priority')
    };
    
    const dueDate = formData.get('due_date');
    if (dueDate) {
        taskData.due_date = new Date(dueDate).toISOString();
    }
    
    try {
        if (currentEditingTask) {
            await updateTask(currentEditingTask.id, taskData);
        } else {
            await createTask(taskData);
        }
        
        closeModal();
    } catch (error) {
        showToast('Failed to save task', 'error');
    }
}

// Event Handlers
function editTask(taskId) {
    const task = tasks.find(t => t.id === taskId);
    if (task) {
        openModal(task);
    }
}

async function confirmDeleteTask(taskId) {
    if (confirm('Are you sure you want to delete this task?')) {
        try {
            await deleteTask(taskId);
        } catch (error) {
            showToast('Failed to delete task', 'error');
        }
    }
}

function attachEventListeners() {
    // Modal controls
    elements.addTaskBtn.addEventListener('click', () => openModal());
    elements.closeModal.addEventListener('click', closeModal);
    elements.cancelBtn.addEventListener('click', closeModal);
    
    // Form submission
    elements.taskForm.addEventListener('submit', handleFormSubmit);
    
    // Close modal on outside click
    elements.taskModal.addEventListener('click', (e) => {
        if (e.target === elements.taskModal) {
            closeModal();
        }
    });
    
    // Keyboard shortcuts
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && elements.taskModal.classList.contains('active')) {
            closeModal();
        }
        
        if (e.key === 'n' && e.ctrlKey) {
            e.preventDefault();
            openModal();
        }
    });
}

// Utility Functions
function showToast(message, type = 'success') {
    elements.toastMessage.textContent = message;
    elements.toast.classList.add('show');
    
    const icon = elements.toast.querySelector('i');
    icon.className = type === 'success' ? 'fas fa-check-circle' : 'fas fa-exclamation-circle';
    
    setTimeout(() => {
        elements.toast.classList.remove('show');
    }, 3000);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Update active users (simulated for now)
setInterval(async () => {
    try {
        const response = await fetch(`${API_BASE_URL}/ws/stats`);
        if (response.ok) {
            const stats = await response.json();
            elements.activeUsers.querySelector('span').textContent = 
                `${stats.total_clients || 0} active`;
        }
    } catch (error) {
        console.error('Failed to fetch stats:', error);
    }
}, 10000);

// Expose functions to global scope for inline onclick handlers
window.editTask = editTask;
window.confirmDeleteTask = confirmDeleteTask;