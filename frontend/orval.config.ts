import { defineConfig } from 'orval';

export default defineConfig({
  mnemosyne: {
    input: {
      target: '../api/openapi.yaml',
      // The checked-in contract uses relative references across api/.
      parserOptions: { externalRefs: { allow: ['*'] } },
    },
    output: {
      target: './src/api/generated/mnemosyne.ts',
      schemas: './src/api/generated/models',
      client: 'react-query',
      httpClient: 'fetch',
      clean: true,
      override: {
        fetch: { includeHttpResponseReturnType: false },
        mutator: {
          path: './src/api/request.ts',
          name: 'request',
        },
      },
    },
  },
});
