import { defineConfig } from 'vitest/config'
import { svelte } from '@sveltejs/vite-plugin-svelte'
// This is annoying, we can just ignore it though
// @ts-ignore ts(1259)
import path from 'node:path';


export default defineConfig({
  plugins: [svelte({hot: !process.env.VITEST})],
  resolve: {
    alias: {
      src: path.resolve('./src')
    }
  },
  test: {
    include: ['src/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'],
    globals: true,
    environment: 'jsdom',
    setupFiles: 'vitest-setup.ts',
    mockReset: true
  },
})