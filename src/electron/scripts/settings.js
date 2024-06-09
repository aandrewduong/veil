document.addEventListener('DOMContentLoaded', async () => {
    const webhookUrlInput = document.getElementById('webhook-url');

    // Load webhook URL from app data
    const webhookUrl = await window.electron.getWebhookUrl();
    webhookUrlInput.value = webhookUrl;

    document.querySelector('.save-settings-btn').addEventListener('click', async function() {
        const webhookUrl = webhookUrlInput.value;
        await window.electron.saveWebhookUrl(webhookUrl);
        window.electron.showToast('Settings saved!');
    });
});