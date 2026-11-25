import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

import { loadAppConfig } from './app-config.ts';

export default defineConfig({
  plugins: [react()],
  define: {
    __MNEMOSYNE_API_PREFIX__: JSON.stringify(loadAppConfig().apiPrefix),
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    restoreMocks: true,
    unstubGlobals: true,
    clearMocks: true,
  },
});
