document.addEventListener('DOMContentLoaded', async () => {
    document.querySelector('.create-btn').addEventListener('click', openCreateModal);

    const deleteButton = document.getElementById('deleteTasks');
    const selectAllCheckbox = document.getElementById('selectAll');

    deleteButton.addEventListener('click', deleteSelectedTasks);
    selectAllCheckbox.addEventListener('change', toggleSelectAll);

    await fetchTerms();
    await loadTasks();
});

// Store term descriptions
let termDescriptions = {};

// Fetch terms from the API and populate dropdown options
async function fetchTerms() {
    const response = await fetch("https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/classSearch/getTerms?searchTerm=&offset=1&max=15");
    const data = await response.json();
    data.forEach(term => {
        termDescriptions[term.code] = term.description;
    });
    addOptions('createTerm', data);
}

// Load tasks from the server and populate the table
async function loadTasks() {
    const response = await fetch(`http://localhost:1942/tasks/all`);
    const tasks = await response.json();
    tasks.forEach(task => {
        const termDescription = termDescriptions[task.term] || task.term;
        addTask(task.id, task.mode, termDescription, task.crns, task.status);
    });
}

// Add a task to the table
function addTask(taskId, mode, term, crns, status) {
    const tableBody = document.querySelector('table tbody');
    const newRow = document.createElement('tr');

    newRow.innerHTML = `
        <td class="checkbox-cell"><input type="checkbox"></td>
        <td>${taskId}</td>
        <td>${mode}</td>
        <td>${term}</td>
        <td>${crns}</td>
        <td>${status}</td>
        <td class="table-actions">
            <img src="./img/play.png" alt="Start" class="start-img">
            <img src="./img/trash.png" alt="Delete" class="delete-img">
        </td>
    `;

    const startImg = newRow.querySelector('.start-img');
    const deleteImg = newRow.querySelector('.delete-img');
    const statusCell = newRow.children[5];

    deleteImg.addEventListener('click', async () => deleteTask(newRow, taskId));
    startImg.addEventListener('click', async () => toggleTaskStatus(startImg, statusCell, taskId));

    tableBody.appendChild(newRow);

    if (status !== 'Stopped') {
        startImg.src = './img/stop.png';
        startImg.alt = 'Stop';
        pollTaskStatus(taskId, startImg, statusCell);
    }
}

// Delete the selected tasks
function deleteSelectedTasks() {
    const selectedTasks = document.querySelectorAll('input[type="checkbox"]:checked');
    selectedTasks.forEach(task => {
        const taskRow = task.closest('tr');
        if (taskRow && task.id !== 'selectAll') {
            taskRow.remove();
        }
    });
}

// Toggle select all checkboxes
function toggleSelectAll() {
    const checkboxes = document.querySelectorAll('tbody input[type="checkbox"]');
    checkboxes.forEach(checkbox => {
        checkbox.checked = this.checked;
    });
}

// Poll the task status and update the table
async function pollTaskStatus(taskId, startImg, statusCell) {
    let pollInterval = setInterval(async () => {
        const newStatus = await getTaskStatus(taskId);
        statusCell.textContent = newStatus;
        if (newStatus === 'Task not found' || newStatus === 'Completed') {
            clearInterval(pollInterval);
            startImg.src = './img/play.png';
            startImg.alt = 'Start';
        }
    }, 1000);
}

// Toggle the task status between running and stopped
async function toggleTaskStatus(startImg, statusCell, taskId) {
    if (statusCell.textContent === 'Stopped') {
        const response = await fetch(`http://localhost:1942/tasks/run?id=${taskId}`);
        const data = await response.json();
        if (data.message === "Task is running") {
            statusCell.textContent = 'Running';
            startImg.src = './img/stop.png';
            startImg.alt = 'Stop';
            pollTaskStatus(taskId, startImg, statusCell);
        }
    } else {
        const response = await fetch(`http://localhost:1942/tasks/delete?id=${taskId}`);
        const data = await response.json();
        if (data.message === 'Task deleted') {
            statusCell.textContent = 'Stopped';
            startImg.src = './img/play.png';
            startImg.alt = 'Start';
        }
    }
}

// Delete a task from the table and server
async function deleteTask(row, taskId) {
    row.remove();
    const response = await fetch(`http://localhost:1942/tasks/delete?id=${taskId}`);
    const data = await response.json();
}

// Fetch the status of a specific task
async function getTaskStatus(taskId) {
    const response = await fetch(`http://localhost:1942/tasks/status?id=${taskId}`);
    const data = await response.json();
    return data.status;
}

// Open the create modal and set default values
function openCreateModal() {
    const modal = document.getElementById('createModal');
    modal.style.display = 'block';

    const createTaskId = document.getElementById('createTaskId');
    const createMode = document.getElementById('createMode');
    const createTerm = document.getElementById('createTerm');
    const createCrns = document.getElementById('createCrns');
    const createCrnsError = document.getElementById('createCrnsError');

    createTaskId.value = Math.random().toString(36).slice(2, 9);
    createMode.value = 'Signup';
    createTerm.value = '';
    createCrns.value = '';

    document.getElementById('createTask').onclick = async () => {
        if (!createCrns.value.trim()) {
            showError(createCrns, createCrnsError, 'CRNs cannot be empty.');
            return;
        }
        const taskId = createTaskId.value;
        const mode = createMode.value;
        const term = createTerm.value;
        const crns = createCrns.value;
        modal.style.display = 'none';
        clearError(createCrns, createCrnsError);

        const credentials = await window.electron.getCredentials();
        const response = await fetch('http://localhost:1942/tasks/create', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: taskId, mode, term, crns, username: credentials.username, password: credentials.password, status: "Created" })
        });
        const data = await response.json();
        if (data.message === 'Task created') {
            addTask(taskId, mode, createTerm.selectedOptions[0].text, crns, 'Stopped');
        }
    };

    setUpModalCloseHandlers(modal, createCrns, createCrnsError);
}

// Set up handlers to close the modal and clear errors
function setUpModalCloseHandlers(modal, inputElement, errorElement) {
    const span = modal.getElementsByClassName('close')[0];
    span.onclick = () => {
        modal.style.display = 'none';
        clearError(inputElement, errorElement);
    };

    window.onclick = (event) => {
        if (event.target == modal) {
            modal.style.display = 'none';
            clearError(inputElement, errorElement);
        }
    };
}

// Show an error message for input validation
function showError(input, errorElement, message) {
    input.classList.add('error');
    errorElement.textContent = message;
    errorElement.style.display = 'block';
}

// Clear an error message
function clearError(input, errorElement) {
    input.classList.remove('error');
    errorElement.textContent = '';
    errorElement.style.display = 'none';
}

// Add options to a select element
function addOptions(elementId, options) {
    const selectElement = document.getElementById(elementId);
    options.forEach(option => {
        const newOption = document.createElement('option');
        newOption.value = option.code;
        newOption.text = option.description;
        selectElement.appendChild(newOption);
    });
}