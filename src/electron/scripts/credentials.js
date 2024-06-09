document.addEventListener('DOMContentLoaded', async () => {
    const usernameInput = document.getElementById('fhdausername');
    const passwordInput = document.getElementById('fhdapassword');

    // Load credentials from app data
    const credentials = await window.electron.getCredentials();
    usernameInput.value = credentials.username;
    passwordInput.value = credentials.password;

    document.querySelector('.save-credentials-btn').addEventListener('click', async function() {
        const username = usernameInput.value;
        const password = passwordInput.value;
        await window.electron.saveCredentials({ username, password });
        window.electron.showToast('Credentials saved!');
    });
});
