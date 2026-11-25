import { readFileSync } from 'node:fs';
import { describe, expect, it, vi } from 'vitest';

import { loadAppConfig } from './app-config.ts';

vi.mock('node:fs', () => ({ readFileSync: vi.fn() }));

describe('shared app config', () => {
  it('reads a custom API prefix and bind address from YAML', () => {
    vi.mocked(readFileSync).mockReturnValue(
      'bindAddress: ":9090"\napiPrefix: /music/v1/\nfrontendDir: /private/assets\n',
    );
    expect(loadAppConfig()).toEqual({
      apiPrefix: '/music/v1/',
      bindAddress: ':9090',
    });
  });

  it.each([
    '/',
    'api/',
    '/api',
    '/api//',
    '/api//v1/',
    '/{path}/',
    '/api?query/',
  ])('rejects unsafe prefix %s', (prefix) => {
    vi.mocked(readFileSync).mockReturnValue(
      `bindAddress: ":8080"\napiPrefix: "${prefix}"\n`,
    );
    expect(loadAppConfig).toThrow('Invalid API prefix or bind address');
  });

  it('rejects missing settings', () => {
    vi.mocked(readFileSync).mockReturnValue('{}');
    expect(loadAppConfig).toThrow('Invalid API prefix or bind address');
  });
});
