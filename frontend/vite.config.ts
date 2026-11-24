import react from '@vitejs/plugin-react';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const proxy = {
    '/api': {
      target: env['BACKEND_URL'] ?? 'http://localhost:8080',
      changeOrigin: true,
      rewrite: (path: string): string => path.replace(/^\/api/, ''),
    },
  };

  return {
    plugins: [react()],
    server: { port: 5173, strictPort: true, proxy },
    preview: { proxy },
  };
});
