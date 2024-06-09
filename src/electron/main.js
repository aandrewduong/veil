const { app, BrowserWindow, shell, ipcMain } = require('electron');
const path = require('path');
const fs = require('fs');
const { spawn } = require('child_process');

let mainWindow;
let subprocess;

// Function to register IPC handlers
function registerIpcHandlers() {
  ipcMain.handle('get-webhook-url', async () => {
    const settingsPath = path.join(app.getPath('userData'), 'settings.json');
    if (fs.existsSync(settingsPath)) {
      const data = fs.readFileSync(settingsPath, 'utf8');
      const settings = JSON.parse(data);
      return settings.webhookUrl || '';
    }
    return '';
  });

  ipcMain.handle('save-webhook-url', async (event, webhookUrl) => {
    const settingsPath = path.join(app.getPath('userData'), 'settings.json');
    const settings = { webhookUrl };
    fs.writeFileSync(settingsPath, JSON.stringify(settings, null, 2));
  });

  ipcMain.handle('get-credentials', async () => {
    const settingsPath = path.join(app.getPath('userData'), 'settings.json');
    if (fs.existsSync(settingsPath)) {
      const data = fs.readFileSync(settingsPath, 'utf8');
      const settings = JSON.parse(data);
      return {
        username: settings.fhdaUsername || '',
        password: settings.fhdaPassword || '',
      };
    }
    return { username: '', password: '' };
  });

  ipcMain.handle('save-credentials', async (event, credentials) => {
    const settingsPath = path.join(app.getPath('userData'), 'settings.json');
    let settings = {};
    if (fs.existsSync(settingsPath)) {
      const data = fs.readFileSync(settingsPath, 'utf8');
      settings = JSON.parse(data);
    }
    settings.fhdaUsername = credentials.username;
    settings.fhdaPassword = credentials.password;
    fs.writeFileSync(settingsPath, JSON.stringify(settings, null, 2));
  });
}

// Function to create the main window
function createWindow() {
  // Create the browser window
  mainWindow = new BrowserWindow({
    width: 1175,
    height: 670,
    titleBarStyle: 'hiddenInset',
    frame: false,
    webPreferences: {
      preload: path.join(__dirname, './scripts/preload.js'),
    },
  });

  mainWindow.loadFile(path.join(__dirname, 'index.html'));

  mainWindow.on('closed', function () {
      if (subprocess) {
          subprocess.kill();
      }
      mainWindow = null;
  });

  const exePath = path.join(process.resourcesPath, 'engine.exe');

  // Log the path to verify it
  console.log(`Engine Path: ${exePath}`);

  // Ensure the engine file exists at the specified location
  if (!fs.existsSync(exePath)) {
      console.error('engine.exe does not exist at the specified path:', exePath);
      return;
  }

  // Spawn the subprocess from the specified location
  subprocess = spawn(exePath);

  subprocess.stdout.on('data', (data) => {
      console.log(`stdout: ${data}`);
  });

  subprocess.stderr.on('data', (data) => {
      console.error(`stderr: ${data}`);
  });

  subprocess.on('close', (code) => {
      console.log(`child process exited with code ${code}`);
      subprocess = null;
  });

  // Load the index.html of the app
  mainWindow.loadFile(path.join(__dirname, 'index.html'));
  mainWindow.setMenuBarVisibility(false);

  // Prevent new windows from opening and open them externally instead
  mainWindow.webContents.on('new-window', (event, url) => {
    event.preventDefault();
    shell.openExternal(url);
  });

  // Minimize and close window handlers
  ipcMain.on('minimize-window', () => {
    mainWindow.minimize();
  });

  ipcMain.on('close-window', () => {
    mainWindow.close();
  });

  // Optionally open DevTools
  // mainWindow.webContents.openDevTools();
}

// App ready event handler
app.whenReady().then(() => {
  registerIpcHandlers();
  createWindow();
});

// Re-create a window in the app when the dock icon is clicked and there are no other windows open (macOS)
app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) createWindow();
});

// Quit when all windows are closed, except on macOS
app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});

// In this file, you can include the rest of your app's specific main process
// code. You can also put them in separate files and require them here.
