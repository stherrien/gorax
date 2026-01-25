import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8181',
        changeOrigin: true,
      },
      // Note: Do NOT proxy /webhooks as it's a frontend route
      // Only /api/v1/webhooks should be proxied (covered by /api above)
    },
  },
  build: {
    // Generate source maps for production error tracking
    sourcemap: true,
    // Reduce chunk size warnings threshold
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        // Code splitting: separate vendor chunks
        manualChunks: {
          // React core
          'vendor-react': ['react', 'react-dom', 'react-router-dom'],
          // Data fetching and state
          'vendor-query': ['@tanstack/react-query', 'zustand'],
          // UI libraries
          'vendor-ui': ['@xyflow/react', 'recharts'],
        },
      },
    },
  },
})
