import type { ReactElement } from 'react';

import { useGetListens } from '../../api/generated/mnemosyne';

export function Listens(): ReactElement {
  const { data, error, isPending, isError, refetch, isFetching } =
    useGetListens();

  return (
    <section aria-labelledby="listens-heading">
      <h1 id="listens-heading">Listens</h1>
      <p>Response from GET /listens</p>
      {isPending && <p role="status">Loading listens…</p>}
      {isError && (
        <div role="alert">
          <p>
            Could not load listens.
            {error instanceof Error ? ` ${error.message}` : ''}
          </p>
          <button
            type="button"
            disabled={isFetching}
            onClick={() => {
              void refetch();
            }}
          >
            Try again
          </button>
        </div>
      )}
      {data !== undefined && (
        <>
          {data.listens.length === 0 && <p>No listens yet.</p>}
          <pre aria-label="Listens response">
            {JSON.stringify(data, null, 2)}
          </pre>
        </>
      )}
    </section>
  );
}
