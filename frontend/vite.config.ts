import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    dedupe: [
      '@fullcalendar/core',
      '@fullcalendar/common',
      '@fullcalendar/daygrid',
      '@fullcalendar/timegrid',
      '@fullcalendar/interaction',
      '@fullcalendar/react',
    ],
  },
  optimizeDeps: {
    exclude: [
      '@fullcalendar/core',
      '@fullcalendar/common',
      '@fullcalendar/daygrid',
      '@fullcalendar/timegrid',
      '@fullcalendar/interaction',
      '@fullcalendar/react',
    ],
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
