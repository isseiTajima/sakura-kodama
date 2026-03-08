import { test, expect } from '@playwright/test'

test('キャラクター画像が表示される', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('img[alt="キャラクター"]')).toBeVisible()
})

test('ツールバーボタンが表示される', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('.toolbar button').first()).toBeVisible()
})

test('#app に内容がある', async ({ page }) => {
  await page.goto('/')
  const content = await page.$eval('#app', el => el.innerHTML)
  expect(content.trim()).not.toBe('')
})
