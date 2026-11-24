# Frontend Architecture & Portability Rules (Vite → Next.js Ready)

This repository is built with **React + Vite**, structured so it can migrate to **Next.js (App Router)** with minimal refactoring. All contributors and AI coding agents must follow these rules.

## 1. Directory & Separation of Concerns

* **`src/components/` & `src/features/` (100% Portable):** Must contain pure React UI components, business logic, and TanStack Query hooks. Never import routing libraries or build-tool globals here.
* **`src/pages/` & `src/router/` (Framework Glue Only):** Keep route definitions, URL parameter extraction (`useParams`, `useSearchParams`), and layout wrappers strictly isolated here. Pages should be thin shells that pass URL params as props into `src/features/` components.

## 2. Routing & Navigation Isolation

* **No direct router imports in components:** Never import `react-router-dom` or `@tanstack/react-router` inside `src/components/` or `src/features/`.
* **Use wrapper components:** Create a single `src/components/AppLink.tsx` wrapper for internal links and a `useAppNavigate()` hook for programmatic redirects. When migrating to Next.js, only these two files need to change.

## 3. Data Fetching & OpenAPI Client Rules

* **TanStack Query only:** All API calls must go through generated OpenAPI TanStack Query hooks (e.g., via Orval or `@hey-api/openapi-ts`). Never call `fetch()` or `axios` directly inside components.
* **No build-tool globals in API clients:** Configure your OpenAPI client's `baseURL` inside a single `src/lib/env.ts` file rather than calling `import.meta.env` across the codebase.

## 4. Environment Variables

* **Centralize env access:** Never reference `import.meta.env.VITE_*` directly inside components, hooks, or utilities.
* **Single adapter file:** Export all environment variables from `src/lib/env.ts` so they can be swapped to `process.env.NEXT_PUBLIC_*` in one place later.

## 5. Browser APIs & Asset Handling

* **Guard browser globals:** Avoid accessing `window`, `document`, or `localStorage` during the top-level render phase of any component. Put browser storage and DOM access inside `useEffect` hooks or event handlers so components remain SSR-safe.
* **Standard `<img>` tags & CSS:** Use standard HTML `<img>` elements (or a single `src/components/AppImage.tsx` wrapper) and standard CSS/Tailwind classes. Avoid Vite-specific asset query imports like `?raw` or `?url`.
