import { chromium, firefox, webkit } from '@playwright/test'
import fs from 'fs'

async function diagnose() {
  const url = process.env.E2E_BASE_URL ?? 'http://127.0.0.1:5173'
  const browserName = (process.env.E2E_BROWSER ?? 'chromium') as 'chromium' | 'firefox' | 'webkit'
  const browserType = { chromium, firefox, webkit }[browserName] ?? chromium
  const browser = await browserType.launch()
  const page = await browser.newPage()

  const consoleErrors: string[] = []
  const consoleAll: string[] = []

  page.on('console', msg => {
    const line = `[${msg.type()}] ${msg.text()}`
    consoleAll.push(line)
    if (msg.type() === 'error' || msg.type() === 'warning') consoleErrors.push(line)
  })
  page.on('pageerror', err => consoleErrors.push(`[pageerror] ${err.message}`))

  await page.goto(url, { waitUntil: 'networkidle', timeout: 15000 })

  fs.mkdirSync('tests/e2e/results', { recursive: true })
  await page.screenshot({ path: 'tests/e2e/results/screenshot.png', fullPage: true })

  const appHTML = await page.$eval('#app', el => el.innerHTML).catch(() => 'EMPTY')
  const dom = await page.content()
  fs.writeFileSync('tests/e2e/results/dom.html', dom)

  const result = {
    timestamp: new Date().toISOString(),
    url,
    browser: browserName,
    hasAppContent: appHTML !== 'EMPTY' && appHTML.trim() !== '',
    consoleErrors,
    consoleAll,
    appHTML: appHTML.slice(0, 3000),
  }
  fs.writeFileSync('tests/e2e/results/diagnostic.json', JSON.stringify(result, null, 2))

  console.log(JSON.stringify(result, null, 2))
  await browser.close()
}

diagnose().catch(err => { console.error(err); process.exit(1) })
