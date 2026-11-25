import { readFileSync } from 'node:fs';
import { parse } from 'yaml';

export interface AppConfig {
  apiPrefix: string;
  bindAddress: string;
}

export function loadAppConfig(): AppConfig {
  const config: unknown = parse(
    readFileSync(
      new URL('../cmd/mnemosyne/config/config.yaml', import.meta.url),
      'utf8',
    ),
  );
  if (
    typeof config !== 'object' ||
    config === null ||
    !('apiPrefix' in config) ||
    typeof config.apiPrefix !== 'string' ||
    !/^(\/[A-Za-z0-9_-]+)+\/$/.test(config.apiPrefix) ||
    !('bindAddress' in config) ||
    typeof config.bindAddress !== 'string'
  ) {
    throw new Error('Invalid API prefix or bind address in config.yaml');
  }
  return { apiPrefix: config.apiPrefix, bindAddress: config.bindAddress };
}
