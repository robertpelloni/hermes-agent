import { app, Menu, nativeImage, shell, Tray, type BrowserWindow } from 'electron'

export interface TrayCallbacks {
  getMainWindow: () => BrowserWindow | null | undefined
  getDashboardUrl: () => string | null | undefined
  restartBackend: () => Promise<void> | void
  stopAllServers: () => void
  getIconPath?: () => string | undefined
}

let appTray: Tray | null = null

export function buildTrayMenu(callbacks: TrayCallbacks): Menu {
  const dashboardUrl = callbacks.getDashboardUrl()
  return Menu.buildFromTemplate([
    {
      label: 'Hermes Agent',
      enabled: false,
    },
    {
      label: 'Status: Running',
      enabled: false,
    },
    { type: 'separator' },
    {
      label: 'Open Main Window',
      click: () => {
        const win = callbacks.getMainWindow()
        if (win && !win.isDestroyed()) {
          if (win.isMinimized()) win.restore()
          win.show()
          win.focus()
        }
      },
    },
    {
      label: 'Open Web Dashboard',
      enabled: Boolean(dashboardUrl),
      click: () => {
        const url = callbacks.getDashboardUrl()
        if (url) {
          shell.openExternal(url).catch(() => {})
        }
      },
    },
    {
      label: 'Documentation',
      click: () => {
        shell.openExternal('https://hermes-agent.nousresearch.com').catch(() => {})
      },
    },
    { type: 'separator' },
    {
      label: 'Restart Backend Server',
      click: async () => {
        try {
          await callbacks.restartBackend()
        } catch {
          // ignore error in tray action
        }
      },
    },
    {
      label: 'Quit Hermes & Stop Servers',
      click: () => {
        callbacks.stopAllServers()
        app.quit()
      },
    },
  ])
}

export function createAppTray(callbacks: TrayCallbacks): Tray | null {
  if (appTray && !appTray.isDestroyed()) {
    return appTray
  }

  const iconPath = callbacks.getIconPath?.()
  let trayImage: Electron.NativeImage
  if (iconPath) {
    try {
      trayImage = nativeImage.createFromPath(iconPath)
      if (process.platform === 'darwin') {
        trayImage = trayImage.resize({ width: 16, height: 16 })
        trayImage.setTemplateImage?.(true)
      } else {
        trayImage = trayImage.resize({ width: 24, height: 24 })
      }
    } catch {
      trayImage = nativeImage.createEmpty()
    }
  } else {
    trayImage = nativeImage.createEmpty()
  }

  try {
    appTray = new Tray(trayImage)
    appTray.setToolTip('Hermes Agent')
    const contextMenu = buildTrayMenu(callbacks)
    appTray.setContextMenu(contextMenu)

    appTray.on('double-click', () => {
      const win = callbacks.getMainWindow()
      if (win && !win.isDestroyed()) {
        if (win.isMinimized()) win.restore()
        win.show()
        win.focus()
      }
    })

    return appTray
  } catch {
    return null
  }
}

export function destroyAppTray(): void {
  if (appTray && !appTray.isDestroyed()) {
    try {
      appTray.destroy()
    } catch {
      // ignore
    }
  }
  appTray = null
}

export function getAppTrayInstance(): Tray | null {
  return appTray
}
