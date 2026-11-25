import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

import { loadAppConfig } from './app-config.ts';

const config = loadAppConfig();
const address = config.bindAddress.startsWith(':')
  ? `localhost${config.bindAddress}`
  : config.bindAddress;
const proxy = {
  [`^${config.apiPrefix}`]: {
    target: `http://${address}`,
    changeOrigin: true,
    rewrite: (path: string): string => path.slice(config.apiPrefix.length - 1),
  },
};

export default defineConfig({
  plugins: [react()],
  define: { __MNEMOSYNE_API_PREFIX__: JSON.stringify(config.apiPrefix) },
  server: { port: 5173, strictPort: true, proxy },
  preview: { proxy },
});
