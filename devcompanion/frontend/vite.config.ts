import fs from 'fs'
import path from 'path'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vite'

const wailsRoot = path.resolve(__dirname, '../wailsjs')
const wailsStubRoot = path.resolve(__dirname, 'src/lib/wails-stub')
const hasWailsBindings = fs.existsSync(wailsRoot)

export default defineConfig({
  plugins: [svelte()],
  root: '.',
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
      wailsjs: hasWailsBindings ? wailsRoot : wailsStubRoot,
    },
  },
  server: {
    host: '127.0.0.1',
    port: 5173,
    fs: {
      allow: ['..'],
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  test: {
    environment: 'jsdom',
    include: ['tests/**/*.test.ts'],
  },
})
