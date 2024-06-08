const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electron', {
    showToast: (message) => {
        const toast = document.createElement('div');
        toast.className = 'toast show';
        
        const toastIcon = document.createElement('div');
        toastIcon.className = 'toast-icon';
        toastIcon.innerHTML = '&#10004;';
        
        const toastMessage = document.createElement('div');
        toastMessage.className = 'toast-message';
        toastMessage.innerText = message;
        
        const toastProgress = document.createElement('div');
        toastProgress.className = 'toast-progress';
        
        const toastClose = document.createElement('div');
        toastClose.className = 'toast-close';
        toastClose.innerHTML = '&#10006;';
        toastClose.onclick = () => closeToast(toast);
        
        toast.appendChild(toastIcon);
        toast.appendChild(toastMessage);
        toast.appendChild(toastProgress);
        toast.appendChild(toastClose);
        
        document.body.appendChild(toast);
        
        setTimeout(() => { closeToast(toast); }, 3000); // Hide after 3 seconds

        function closeToast(toastElement) {
            toastElement.classList.remove('show');
            setTimeout(() => { toastElement.remove(); }, 500); // Remove element after transition
        }
    },
    minimizeWindow: () => ipcRenderer.send('minimize-window'),
    closeWindow: () => ipcRenderer.send('close-window'),
    openDocumentation: () => {
    },
    getSupport: () => {
    }
});

window.addEventListener('DOMContentLoaded', () => {
    const replaceText = (selector, text) => {
        const element = document.getElementById(selector);
        if (element) element.innerText = text;
    }

    for (const type of ['chrome', 'node', 'electron']) {
        replaceText(`${type}-version`, process.versions[type]);
    }
});

async function updateConnectionStatus() {
    const statusIndicator = document.getElementById('statusIndicator');
    const statusText = document.getElementById('statusText');
    try {
        const response = await fetch('http://localhost:1942/status');
        const data = await response.json();
        
        if (data.status === 'Connected') {
            statusIndicator.classList.remove('disconnected');
            statusIndicator.classList.add('connected');
            statusText.textContent = 'Connected';
        }
    } catch (error) {
        statusIndicator.classList.remove('connected');
        statusIndicator.classList.add('disconnected');
        statusText.textContent = 'Disconnected';
    }
}

document.addEventListener('DOMContentLoaded', function() {
    updateConnectionStatus();
    setInterval(updateConnectionStatus, 5000);
});