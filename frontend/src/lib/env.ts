// Framework-specific environment access stays in this adapter.
export const env = {
  apiBaseUrl:
    (import.meta.env['VITE_API_BASE_URL'] as string | undefined) ?? '/api',
};
