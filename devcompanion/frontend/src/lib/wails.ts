import {
  LoadConfig as nativeLoadConfig,
  SaveConfig as nativeSaveConfig,
  OnCharaClick as nativeOnCharaClick,
  InstallOllama as nativeInstallOllama,
  CancelInstall as nativeCancelInstall,
  DetectSetupStatus as nativeDetectSetupStatus,
  CompleteSetup as nativeCompleteSetup,
  ExpandForOnboarding as nativeExpandForOnboarding,
} from 'wailsjs/go/main/App'

type RuntimeAwareWindow = Window & {
  runtime?: unknown
  process?: { env?: Record<string, string | undefined> }
}

type GlobalProcess = { env?: Record<string, string | undefined> }

export type AppConfig = {
  name: string
  user_name: string
  tone: string
  encourage_freq: string
  monologue: boolean
  always_on_top: boolean
  click_through: boolean
  mute: boolean
  scale: number
  log_path: string
  model: string
  ollama_endpoint: string
  independent_window_opacity: number
  llm_backend: string
  anthropic_api_key: string
  setup_completed: boolean
  auto_start: boolean
  speech_frequency: number
}

const DEFAULT_NAME = 'サクラ'
const DEFAULT_MODEL = 'gemma3:4b'
const DEFAULT_OLLAMA_ENDPOINT = 'http://localhost:11434/api/generate'
const CLAUDE_LOG_SEGMENTS = ['Library', 'Logs', 'Claude', 'Claude Code.log'] as const

export const defaultConfig: AppConfig = {
  name: DEFAULT_NAME,
  user_name: '',
  tone: 'genki',
  encourage_freq: 'mid',
  monologue: true,
  always_on_top: true,
  click_through: false,
  mute: false,
  scale: 1,
  log_path: resolveDefaultLogPath(),
  model: DEFAULT_MODEL,
  ollama_endpoint: DEFAULT_OLLAMA_ENDPOINT,
  independent_window_opacity: 1,
  llm_backend: 'router',
  anthropic_api_key: '',
  setup_completed: false,
  auto_start: true,
  speech_frequency: 2,
}

const FALLBACK_STORAGE_KEY = 'wails:fallbackConfig'
let fallbackConfigCache: AppConfig | null = null

function resolveHomeDirectory(): string | null {
  if (typeof window !== 'undefined') {
    const home = (window as RuntimeAwareWindow).process?.env?.HOME
    if (home && home.trim()) {
      return home.trim()
    }
  }
  const globalProcess = (globalThis as typeof globalThis & { process?: GlobalProcess }).process
  const home = globalProcess?.env?.HOME
  if (home && home.trim()) {
    return home.trim()
  }
  return null
}

function resolveDefaultLogPath(): string {
  const home = resolveHomeDirectory()
  if (!home) {
    return ''
  }
  const normalizedHome = home.replace(/[/\\]+$/, '')
  const separator = home.includes('\\') ? '\\' : '/'
  return [normalizedHome, ...CLAUDE_LOG_SEGMENTS].join(separator)
}

function hasRuntime(): boolean {
  return typeof window !== 'undefined' && Boolean((window as RuntimeAwareWindow).runtime)
}

function cloneDefaultConfig(partial?: Partial<AppConfig>): AppConfig {
  const baseLogPath = defaultConfig.log_path || resolveDefaultLogPath()
  return { ...defaultConfig, log_path: baseLogPath, ...partial }
}

function readStoredFallback(): Partial<AppConfig> | null {
  if (typeof window === 'undefined' || !window.localStorage) {
    return null
  }
  try {
    const value = window.localStorage.getItem(FALLBACK_STORAGE_KEY)
    if (!value) {
      return null
    }
    const parsed = JSON.parse(value) as Partial<AppConfig> | null
    return parsed && typeof parsed === 'object' ? parsed : null
  } catch {
    return null
  }
}

function writeStoredFallback(config: AppConfig): void {
  if (typeof window === 'undefined' || !window.localStorage) {
    return
  }
  try {
    window.localStorage.setItem(FALLBACK_STORAGE_KEY, JSON.stringify(config))
  } catch {
    // ignore storage failures in fallback mode
  }
}

export async function loadConfig(): Promise<AppConfig> {
  if (!hasRuntime()) {
    if (!fallbackConfigCache) {
      const stored = readStoredFallback()
      fallbackConfigCache = cloneDefaultConfig(stored ?? undefined)
    }
    return { ...fallbackConfigCache }
  }
  const loaded = (await nativeLoadConfig()) as Partial<AppConfig> | undefined
  return cloneDefaultConfig(loaded)
}

export async function saveConfig(config: AppConfig): Promise<void> {
  if (!hasRuntime()) {
    fallbackConfigCache = { ...config }
    writeStoredFallback(fallbackConfigCache)
    return
  }
  await nativeSaveConfig(config)
}

export function onCharaClick(): Promise<unknown> | void {
  if (!hasRuntime()) {
    return Promise.resolve()
  }
  return nativeOnCharaClick()
}

export async function InstallOllama(): Promise<void> {
  if (!hasRuntime()) {
    console.log('InstallOllama triggered (no runtime)')
    return
  }
  await nativeInstallOllama()
}

export async function DetectSetupStatus(): Promise<{is_first_run: boolean, detected_backends: string[], has_claude_key: boolean}> {
  if (!hasRuntime()) {
    return { is_first_run: true, detected_backends: [], has_claude_key: false }
  }
  return nativeDetectSetupStatus()
}

export async function CancelInstall(): Promise<void> {
  if (!hasRuntime()) return
  await nativeCancelInstall()
}

export async function SetClickThrough(enabled: boolean): Promise<void> {
  if (!hasRuntime()) return
  const app = (window as any)?.go?.main?.App
  if (app && typeof app.SetClickThrough === 'function') {
    await app.SetClickThrough(enabled)
  }
}

export async function CompleteSetup(): Promise<void> {
  if (!hasRuntime()) {
    return
  }
  await nativeCompleteSetup()
}

export async function ExpandForOnboarding(): Promise<void> {
  if (!hasRuntime()) {
    return
  }
  await nativeExpandForOnboarding()
}

export function onOpenSettings(cb: () => void) {
  if (typeof window !== 'undefined' && (window as any).runtime) {
    const r = (window as any).runtime
    r.EventsOff('open-settings')
    r.EventsOn('open-settings', cb)
  }
}
