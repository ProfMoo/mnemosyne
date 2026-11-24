import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ListensPage } from '../../pages/ListensPage';

function renderPage(): void {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  render(
    <QueryClientProvider client={client}>
      <ListensPage />
    </QueryClientProvider>,
  );
}

describe('listens page', () => {
  it('shows loading state while the request is pending', () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() => new Promise<Response>(() => undefined)),
    );
    renderPage();
    expect(screen.getByRole('status')).toHaveTextContent('Loading listens');
  });

  it('calls the generated endpoint and displays its response', async () => {
    const body = {
      listens: [{ id: 'listen-1', track: { name: 'Test track' } }],
    };
    const fetchMock = vi.fn().mockResolvedValue(Response.json(body));
    vi.stubGlobal('fetch', fetchMock);
    renderPage();

    expect(await screen.findByLabelText('Listens response')).toHaveTextContent(
      'Test track',
    );
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/listens',
      expect.objectContaining({ method: 'GET' }),
    );
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('handles an empty response', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(Response.json({ listens: [] })),
    );
    renderPage();
    expect(await screen.findByText('No listens yet.')).toBeInTheDocument();
    expect(screen.getByLabelText('Listens response')).toHaveTextContent(
      '"listens": []',
    );
  });

  it('shows HTTP failures and allows retrying', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        Response.json(
          { error: 'unavailable' },
          { status: 500, statusText: 'Internal Server Error' },
        ),
      )
      .mockResolvedValueOnce(Response.json({ listens: [] }));
    vi.stubGlobal('fetch', fetchMock);
    renderPage();

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Request failed: 500',
    );
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
    expect(await screen.findByText('No listens yet.')).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('shows network failures', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockRejectedValue(new TypeError('Failed to fetch')),
    );
    renderPage();
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Failed to fetch',
    );
  });
});
