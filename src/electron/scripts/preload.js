const { contextBridge, ipcRenderer } = require('electron');

// Expose Electron APIs to the renderer process
contextBridge.exposeInMainWorld('electron', {
    showToast: (message) => showCustomToast(message),
    getWebhookUrl: () => ipcRenderer.invoke('get-webhook-url'),
    saveWebhookUrl: (webhookUrl) => ipcRenderer.invoke('save-webhook-url', webhookUrl),
    minimizeWindow: () => ipcRenderer.send('minimize-window'),
    closeWindow: () => ipcRenderer.send('close-window'),
    openDocumentation: () => {}, // Placeholder for future implementation
    getSupport: () => {}, // Placeholder for future implementation
    getCredentials: () => ipcRenderer.invoke('get-credentials'),
    saveCredentials: (credentials) => ipcRenderer.invoke('save-credentials', credentials),
});

// Show a custom toast notification
function showCustomToast(message) {
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

    // Hide the toast after 3 seconds
    setTimeout(() => { closeToast(toast); }, 3000);

    // Close and remove the toast element
    function closeToast(toastElement) {
        toastElement.classList.remove('show');
        setTimeout(() => { toastElement.remove(); }, 500);
    }
}

// Update connection status with the server
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

// DOMContentLoaded event listener to initialize the status update and version display
document.addEventListener('DOMContentLoaded', () => {
    // Replace text content for version info
    const replaceText = (selector, text) => {
        const element = document.getElementById(selector);
        if (element) element.innerText = text;
    }

    // Display versions for Chrome, Node, and Electron
    for (const type of ['chrome', 'node', 'electron']) {
        replaceText(`${type}-version`, process.versions[type]);
    }

    const documentationLink = document.getElementById('documentation-link');

    documentationLink.addEventListener('click', (event) => {
      event.preventDefault();
      const { shell } = require('electron');
      shell.openExternal(documentationLink.href);
    });
    // Update connection status initially and every 5 seconds
    updateConnectionStatus();
    setInterval(updateConnectionStatus, 5000);
});