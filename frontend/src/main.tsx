import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import { ListensPage } from './pages/ListensPage';
import './styles.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
});

const root = document.getElementById('root');

if (root === null) {
  throw new Error('Missing root element');
}

createRoot(root).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ListensPage />
    </QueryClientProvider>
  </StrictMode>,
);
