import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('@univerjs')) {
            if (id.includes('/locale/')) return 'univer-locale'
            if (id.includes('/engine-formula')) return 'univer-formula'
            if (id.includes('/sheets')) return 'univer-sheets'
            if (id.includes('/ui-adapter-vue3')) return 'univer-vue3'
            if (id.includes('/ui/')) return 'univer-ui'
            return 'univer-core'
          }
          if (id.includes('element-plus')) return 'element-plus'
          if (id.includes('/node_modules/')) return 'vendor'
        },
      },
    },
  },
})