const { app, BrowserWindow, Menu, Tray, ipcMain } = require('electron');
const path = require('path');
const fs = require('fs');
const log = require('electron-log');
const net = require('net');

// Configure logging
log.transports.file.level = 'info';
log.info('Starting CloudSnooze GUI...');

// Keep a global reference of the window object to prevent garbage collection
let mainWindow;
let tray;
let isQuitting = false;

// Default socket path
const DEFAULT_SOCKET_PATH = process.platform === 'win32'
  ? '\\\\.\\pipe\\snooze.sock'
  : '/var/run/snooze.sock';

async function createWindow() {
  mainWindow = new BrowserWindow({
    width: 900,
    height: 700,
    title: 'CloudSnooze',
    icon: path.join(__dirname, 'assets', 'icon.png'),
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: false
    }
  });

  // Load the index.html
  await mainWindow.loadFile('index.html');

  // Open DevTools in development
  if (process.env.NODE_ENV === 'development') {
    mainWindow.webContents.openDevTools();
  }

  // Handle window close event
  mainWindow.on('close', (event) => {
    if (!isQuitting) {
      event.preventDefault();
      mainWindow.hide();
      return false;
    }
    return true;
  });

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

// Create system tray icon
function createTray() {
  const iconPath = path.join(__dirname, 'assets', 'tray-icon.png');
  tray = new Tray(iconPath);
  
  const contextMenu = Menu.buildFromTemplate([
    { label: 'Show CloudSnooze', click: () => mainWindow.show() },
    { type: 'separator' },
    { label: 'Start Monitoring', click: () => sendCommandToDaemon('START') },
    { label: 'Stop Monitoring', click: () => sendCommandToDaemon('STOP') },
    { type: 'separator' },
    { label: 'Quit', click: () => {
      isQuitting = true;
      app.quit();
    }}
  ]);
  
  tray.setToolTip('CloudSnooze');
  tray.setContextMenu(contextMenu);
  tray.on('click', () => {
    mainWindow.isVisible() ? mainWindow.hide() : mainWindow.show();
  });
}

// Initialize the app
app.whenReady().then(() => {
  createWindow();
  createTray();
  
  // Check daemon status every 30 seconds
  setInterval(checkDaemonStatus, 30000);
  
  // Initial status check
  checkDaemonStatus();
  
  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

// Quit when all windows are closed, except on macOS
app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

// Set up IPC listeners
ipcMain.on('get-status', (event) => {
  sendCommandToDaemon('STATUS')
    .then(result => {
      event.reply('status-result', result);
    })
    .catch(err => {
      event.reply('status-result', { error: err.message });
    });
});

ipcMain.on('get-config', (event) => {
  sendCommandToDaemon('CONFIG_GET')
    .then(result => {
      event.reply('config-result', result);
    })
    .catch(err => {
      event.reply('config-result', { error: err.message });
    });
});

ipcMain.on('set-config', (event, params) => {
  sendCommandToDaemon('CONFIG_SET', params)
    .then(result => {
      event.reply('config-set-result', result);
    })
    .catch(err => {
      event.reply('config-set-result', { error: err.message });
    });
});

ipcMain.on('get-history', (event, params) => {
  sendCommandToDaemon('HISTORY', params)
    .then(result => {
      event.reply('history-result', result);
    })
    .catch(err => {
      event.reply('history-result', { error: err.message });
    });
});

// Check daemon status
function checkDaemonStatus() {
  sendCommandToDaemon('STATUS')
    .then(result => {
      if (mainWindow) {
        mainWindow.webContents.send('status-update', result);
      }
    })
    .catch(err => {
      log.error('Failed to check daemon status:', err);
      if (mainWindow) {
        mainWindow.webContents.send('daemon-error', { message: err.message });
      }
    });
}

// Send command to daemon via Unix socket
function sendCommandToDaemon(command, params = {}) {
  return new Promise((resolve, reject) => {
    const socketPath = DEFAULT_SOCKET_PATH;
    
    // Check if socket exists
    if (!fs.existsSync(socketPath)) {
      reject(new Error(`Socket not found at ${socketPath}. Is the daemon running?`));
      return;
    }
    
    const client = net.createConnection({ path: socketPath }, () => {
      // Connected to socket
      const request = {
        command: command,
        params: params
      };
      
      client.write(JSON.stringify(request));
    });
    
    let responseData = '';
    
    client.on('data', (data) => {
      responseData += data.toString();
    });
    
    client.on('end', () => {
      try {
        const response = JSON.parse(responseData);
        if (response.status === 'error') {
          reject(new Error(response.error));
        } else {
          resolve(response.data);
        }
      } catch (err) {
        reject(new Error(`Invalid response from daemon: ${err.message}`));
      }
    });
    
    client.on('error', (err) => {
      reject(new Error(`Socket error: ${err.message}`));
    });
  });
}