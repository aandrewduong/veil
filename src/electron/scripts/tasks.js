document.querySelector('.create-btn').addEventListener('click', function() {
    openCreateModal();
});

function addTask(taskId, mode, term, crns, status) {
    console.log(status);
    const tableBody = document.querySelector('table tbody');
    const newRow = document.createElement('tr');

    const checkboxCell = document.createElement('td');
    checkboxCell.className = 'checkbox-cell';
    const checkbox = document.createElement('input');
    checkbox.type = 'checkbox';
    checkboxCell.appendChild(checkbox);
    newRow.appendChild(checkboxCell);

    const taskIdCell = document.createElement('td');
    taskIdCell.textContent = taskId;
    newRow.appendChild(taskIdCell);

    const modeCell = document.createElement('td');
    modeCell.textContent = mode;
    newRow.appendChild(modeCell);

    const termCell = document.createElement('td');
    termCell.textContent = term;
    newRow.appendChild(termCell);

    const crnsCell = document.createElement('td');
    crnsCell.textContent = crns;
    newRow.appendChild(crnsCell);

    const statusCell = document.createElement('td');
    statusCell.textContent = status;
    newRow.appendChild(statusCell);

    const actionsCell = document.createElement('td');
    actionsCell.className = 'table-actions';

    const startImg = document.createElement('img');
    startImg.src = './img/play.png';
    startImg.alt = 'Start';

    const editImg = document.createElement('img');
    editImg.src = './img/pencil.png';
    editImg.alt = 'Edit Icon';

    const deleteImg = document.createElement('img');
    deleteImg.src = './img/trash.png';
    deleteImg.alt = 'Delete Icon';

    const watchImg = document.createElement('img');
    watchImg.src = './img/eye.png';
    watchImg.alt = 'Watch Icon';

    deleteImg.addEventListener('click', async function() {
        const response = await fetch(`http://localhost:1942/tasks/delete?id=${taskId}`);
        const data = await response.json();
        if (data.message == "Task deleted") {
            tableBody.removeChild(newRow);
        }            
    });

    let pollInterval;

    if (status === 'Running') {
        startImg.src = './img/stop.png';
        startImg.alt = 'Stop';
        pollInterval = setInterval(async () => {
            const newStatus = await pollTaskStatus(taskId);
            statusCell.textContent = newStatus;

            if (newStatus === 'Task not found') {
                clearInterval(pollInterval);
            } else if (newStatus === 'Completed') {
                clearInterval(pollInterval);
                startImg.src = './img/play.png';
                startImg.alt = 'Start';
            }
        }, 100);
    }

    async function togglePlayStop() {
        if (statusCell.textContent === 'Stopped') {
            const response = await fetch(`http://localhost:1942/tasks/run?id=${taskId}`);
            const data = await response.json();
            if (data.message == "Task is running") {
                console.log(data);
                statusCell.textContent = 'Running';
                startImg.src = './img/stop.png';
                startImg.alt = 'Stop';
                pollInterval = setInterval(async () => {
                    const newStatus = await pollTaskStatus(taskId);
                    statusCell.textContent = newStatus;
                    if (newStatus === 'Task not found') {
                        clearInterval(pollInterval);
                    } else if (newStatus === 'Completed') {
                        clearInterval(pollInterval);
                        startImg.src = './img/play.png';
                        startImg.alt = 'Start';
                    }
                }, 100);
            }
        } else {
            statusCell.textContent = 'Stopped';
            startImg.src = './img/play.png';
            startImg.alt = 'Start';
            clearInterval(pollInterval);
        }
    }

    startImg.addEventListener('click', togglePlayStop);

    actionsCell.appendChild(startImg);
    actionsCell.appendChild(editImg);
    actionsCell.appendChild(deleteImg);
    actionsCell.appendChild(watchImg);
    newRow.appendChild(actionsCell);

    watchImg.addEventListener('click', function() {
        openWatchModal(taskId);
    })

    editImg.addEventListener('click', function() {
        openEditModal(taskId, mode, term, crns, newRow);
    });

    tableBody.appendChild(newRow);
}

function openWatchModal(taskId) {
    const modal = document.getElementById('watchModal');
    const watchTaskId = document.getElementById('watchTaskId');
    const watchTaskLog = document.getElementById('watchTaskLog');

    watchTaskId.value = taskId;
    watchTaskLog.value = localStorage.getItem(`log_${taskId}`) || '';

    modal.style.display = 'block';

    const span = modal.getElementsByClassName('close')[0];
    span.onclick = function() {
        modal.style.display = 'none';
    }

    window.onclick = function(event) {
        if (event.target == modal) {
            modal.style.display = 'none';
        }
    }
}

function openEditModal(taskId, mode, term, crns, row) {
    const modal = document.getElementById('editModal');
    const span = modal.getElementsByClassName('close')[0];
    const editTaskId = document.getElementById('editTaskId');
    const editMode = document.getElementById('editMode');
    const editTerm = document.getElementById('editTerm');
    const editCrns = document.getElementById('editCrns');
    const editCrnsError = document.getElementById('editCrnsError');

    editTaskId.value = taskId;
    editMode.value = mode;
    editTerm.value = term;
    editCrns.value = crns;

    modal.style.display = 'block';

    span.onclick = function() {
        modal.style.display = 'none';
        clearError(editCrns, editCrnsError);
    }

    window.onclick = function(event) {
        if (event.target == modal) {
            modal.style.display = 'none';
            clearError(editCrns, editCrnsError);
        }
    }

    document.getElementById('saveEdit').onclick = function() {
        if (!editCrns.value.trim()) {
            showError(editCrns, editCrnsError, 'CRNs cannot be empty.');
            return;
        }
        row.children[2].textContent = editMode.value;
        row.children[3].textContent = editTerm.value;
        row.children[4].textContent = editCrns.value;
        modal.style.display = 'none';
        clearError(editCrns, editCrnsError);
    }
}

function openCreateModal() {
    const modal = document.getElementById('createModal');
    const span = modal.getElementsByClassName('close')[0];
    const createTaskId = document.getElementById('createTaskId');
    const createMode = document.getElementById('createMode');
    const createTerm = document.getElementById('createTerm');
    const createCrns = document.getElementById('createCrns');
    const createCrnsError = document.getElementById('createCrnsError');

    createTaskId.value = Math.random().toString(36).slice(2, 7);
    createMode.value = 'Signup';
    createTerm.value = '2024 Fall De Anza';
    createCrns.value = '';

    modal.style.display = 'block';

    span.onclick = function() {
        modal.style.display = 'none';
        clearError(createCrns, createCrnsError);
    }

    window.onclick = function(event) {
        if (event.target == modal) {
            modal.style.display = 'none';
            clearError(createCrns, createCrnsError);
        }
    }

    document.getElementById('createTask').onclick = async function() {
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
        const response = await fetch('http://localhost:1942/tasks/create', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ id: taskId, mode: mode, term, term, crns: crns, status: "Created" })
        });
        const data = await response.json();
        if (data.message == 'Task created') {
            var selectElement = document.getElementById('createTerm');
            var selectedOption = selectElement.selectedOptions[0];
            var selectedText = selectedOption.text;
            addTask(taskId, mode, selectedText, crns, 'Stopped');
        }
    }
}

function showError(input, errorElement, message) {
    input.classList.add('error');
    errorElement.textContent = message;
    errorElement.style.display = 'block';
}

function clearError(input, errorElement) {
    input.classList.remove('error');
    errorElement.textContent = '';
    errorElement.style.display = 'none';
}

async function loadTasks() {
    const response = await fetch(`http://localhost:1942/tasks/all`);
    const data = await response.json();
     data.forEach(task => {
        addTask(task.id, task.mode, task.term, task.crns, task.status);
    });
}

async function pollTaskStatus(Id) {
    const response = await fetch(`http://localhost:1942/tasks/status?id=${Id}`);
    const data = await response.json();
    return data.status;
}

function addOptions(element, options) {
    const selectElement = document.getElementById(element);
    options.forEach(option => {
        // Create a new option element
        const newOption = document.createElement('option');
        console.log(option.code);
        newOption.value = option.code;
        newOption.text = option.description;
        // Append the new option to the select element
        selectElement.appendChild(newOption);
    });
}

window.addEventListener('load', loadTasks);
window.addEventListener('load', function() {
    fetch("https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/classSearch/getTerms?searchTerm=&offset=1&max=10")
    .then((response) => response.json())
    .then((data) => {
        addOptions('createTerm', data);
        addOptions('editTerm', data);
    });
});

document.addEventListener('DOMContentLoaded', function () {
    const deleteButton = document.getElementById('deleteTasks');
    const selectAllCheckbox = document.getElementById('selectAll');

    deleteButton.addEventListener('click', function () {
        const selectedTasks = document.querySelectorAll('input[type="checkbox"]:checked');
        selectedTasks.forEach(task => {
            // Assuming each task row has an ID corresponding to the task ID
            const taskRow = task.closest('tr');
            if (taskRow && task.id !== 'selectAll') {
                taskRow.remove();
            }
        });
    });

    selectAllCheckbox.addEventListener('change', function () {
        const checkboxes = document.querySelectorAll('tbody input[type="checkbox"]');
        checkboxes.forEach(checkbox => {
            checkbox.checked = selectAllCheckbox.checked;
        });
    });
});
