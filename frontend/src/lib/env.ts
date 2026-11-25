// Vite injects this public value from the same YAML config embedded by Go.
declare const __MNEMOSYNE_API_PREFIX__: string;

export const env = { apiBaseUrl: __MNEMOSYNE_API_PREFIX__ };
