import { describe, it, expect, vi, beforeEach } from 'vitest'
import { app } from 'electron'
import { buildTrayMenu } from './tray'

vi.mock('electron', () => {
  return {
    app: {
      quit: vi.fn(),
    },
    Menu: {
      buildFromTemplate: vi.fn((items) => items),
    },
    shell: {
      openExternal: vi.fn().mockResolvedValue(true),
    },
    nativeImage: {
      createFromPath: vi.fn(() => ({
        resize: vi.fn().mockReturnThis(),
        setTemplateImage: vi.fn(),
      })),
      createEmpty: vi.fn(() => ({})),
    },
    Tray: vi.fn().mockImplementation(() => ({
      setToolTip: vi.fn(),
      setContextMenu: vi.fn(),
      on: vi.fn(),
      destroy: vi.fn(),
      isDestroyed: vi.fn(() => false),
    })),
  }
})

describe('electron/tray', () => {
  let mockWin: any
  let callbacks: any

  beforeEach(() => {
    vi.clearAllMocks()
    mockWin = {
      isDestroyed: vi.fn(() => false),
      isMinimized: vi.fn(() => false),
      restore: vi.fn(),
      show: vi.fn(),
      focus: vi.fn(),
    }

    callbacks = {
      getMainWindow: vi.fn(() => mockWin),
      getDashboardUrl: vi.fn(() => 'http://127.0.0.1:9120'),
      restartBackend: vi.fn().mockResolvedValue(undefined),
      stopAllServers: vi.fn(),
      getIconPath: vi.fn(() => '/mock/icon.png'),
    }
  })

  it('builds tray menu with expected items', () => {
    const items = buildTrayMenu(callbacks) as any
    expect(items.length).toBeGreaterThan(5)

    const labels = items.map((i: any) => i.label).filter(Boolean)
    expect(labels).toContain('Hermes Agent')
    expect(labels).toContain('Status: Running')
    expect(labels).toContain('Open Main Window')
    expect(labels).toContain('Open Web Dashboard')
    expect(labels).toContain('Restart Backend Server')
    expect(labels).toContain('Quit Hermes & Stop Servers')
  })

  it('handles Open Main Window click', () => {
    mockWin.isMinimized.mockReturnValue(true)
    const items = buildTrayMenu(callbacks) as any
    const openItem = items.find((i: any) => i.label === 'Open Main Window')
    openItem.click()

    expect(mockWin.restore).toHaveBeenCalled()
    expect(mockWin.show).toHaveBeenCalled()
    expect(mockWin.focus).toHaveBeenCalled()
  })

  it('handles Restart Backend Server click', async () => {
    const items = buildTrayMenu(callbacks) as any
    const restartItem = items.find((i: any) => i.label === 'Restart Backend Server')
    await restartItem.click()

    expect(callbacks.restartBackend).toHaveBeenCalled()
  })

  it('handles Quit Hermes click and stops servers', () => {
    const items = buildTrayMenu(callbacks) as any
    const quitItem = items.find((i: any) => i.label === 'Quit Hermes & Stop Servers')
    quitItem.click()

    expect(callbacks.stopAllServers).toHaveBeenCalled()
    expect(app.quit).toHaveBeenCalled()
  })
})
