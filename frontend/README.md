# Frontend

React + Vite, with generated Orval/TanStack Query hooks from
[`api/openapi.yaml`](../api/openapi.yaml). The initial screen displays the raw
`GET /listens` response, including loading, empty, and retryable error states.
The backend currently returns `500 NOT_IMPLEMENTED` for this endpoint.

## Run The Full App With Docker

From the repository root, with Docker running:

```sh
make docker-start
# Open http://localhost:8080
make docker-stop
```

The multi-stage Dockerfile builds both frontend and Go backend. One Go process
serves the UI and API; Node and Vite are not required at runtime. SQLite persists
in the `mnemosyne-data` Docker volume, which `docker-stop` does not delete.
Stop the existing container before rebuilding with `docker-start`.

Override Make variables if needed:

```sh
make docker-start DOCKER_PORT=8081 DOCKER_CONTAINER=mnemosyne-dev
make docker-stop DOCKER_CONTAINER=mnemosyne-dev
```

## Develop With Live Reload

Install toolchains and dependencies from the repository root:

```sh
mise install
mise exec -- make frontend-install
```

Use two terminals:

```sh
# Terminal 1: Go API on localhost:8080
mise exec -- make run-dev
```

```sh
# Terminal 2: Vite UI on localhost:5173
mise exec -- make frontend-dev
```

Vite reads `apiPrefix` and `bindAddress` from
[`cmd/mnemosyne/config/config.yaml`](../cmd/mnemosyne/config/config.yaml).
By default it proxies `/api/*` to `http://localhost:8080/*`, avoiding development
CORS configuration. Go serves the configured prefix alongside original API paths.

For a local production-style run, build the frontend first:

```sh
mise exec -- make frontend-build
mise exec -- make run-dev
# Open http://localhost:8080
```

Go serves the YAML config's `frontendDir` (`frontend/dist` by default).
Docker embeds `/app/frontend/dist` for this setting during its build.
Production deployment must serve built assets with Go or another static server;
`vite preview` is only a local preview tool.

## Quality Checks

From the repository root:

```sh
mise exec -- make frontend-gen
mise exec -- make frontend-check
```

Or from this directory:

```sh
npm run format         # Apply Prettier
npm run format:check   # Verify formatting
npm run lint           # ESLint + Stylelint, zero warnings
npm run typecheck
npm test
npm run generate:check # Detect drift from the OpenAPI contract
npm run check          # Formatting, linting, types, tests, production build
```

TypeScript enables strict mode, exact optional properties, unchecked-index
checks, unused-code checks, and other compiler safeguards. ESLint uses
`strictTypeChecked` and `stylisticTypeChecked`, React Hooks rules, and React
Refresh rules. CSS uses Stylelint's standard config. Prettier owns formatting.
CI checks generation drift and all quality gates, then builds the Docker image.

Generated sources are committed, type-checked, and excluded from hand-written
lint/format rules. Regenerate after API changes; never edit them manually.
UI code uses generated query hooks, not direct HTTP calls. The shared request
adapter rejects HTTP errors and forwards cancellation signals.

## Structure

- `src/features/`: portable React features and query hooks.
- `src/components/`: shared portable UI, when needed.
- `src/pages/`: thin framework glue.
- `src/api/generated/`: Orval-generated client and models.
- `src/api/request.ts`: shared HTTP transport.
- `src/lib/env.ts`: adapter for the public API prefix injected by Vite.
- `app-config.ts`: reads shared YAML config for Vite and Vitest.

All app settings come from YAML config, not environment variables. `apiPrefix`
is a public build-time setting shared by Go, the browser client, and Vite's proxy.
Use slash-separated letters, digits, underscores, or hyphens (for example,
`/music/api/`), with a required trailing slash. Missing slashes fail validation
rather than being added automatically. Restart Vite after config changes and
rebuild both frontend and backend for deployment. Only the API prefix is bundled
into browser code; private backend settings stay out of the client bundle.

## VS Code

Install workspace-recommended ESLint, Prettier, and Stylelint extensions.
Run dependency installation, then select **TypeScript: Select TypeScript
Version → Use Workspace Version**. Workspace settings enable format-on-save
and ESLint fixes, resolve tools from `frontend/`, and preserve Go tooling.

## Upgrade Dependencies

Versions are pinned exactly in `package.json` and `package-lock.json`.
Node and npm are pinned in [`mise.toml`](../mise.toml).
TypeScript uses the newest compatible 6.x release because current
`typescript-eslint` does not support TypeScript 7.

```sh
npm outdated
npm install --save-dev --save-exact typescript@6
```

Once `typescript-eslint` supports TypeScript 7:

```sh
npm install --save-dev --save-exact typescript@latest typescript-eslint@latest
npm run generate
npm run check
```

When upgrading Node/npm, update `mise.toml`, package engines, Dockerfile, and CI
together. Keep other dependencies on latest compatible stable releases.
