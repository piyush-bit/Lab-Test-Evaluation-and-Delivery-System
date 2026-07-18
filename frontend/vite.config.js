import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: '../TDES/cmd/euc2/ui_dist',
    emptyOutDir: true,
  },
})

