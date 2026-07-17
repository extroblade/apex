# Frontend guide

React 18 + TypeScript, built with **rsbuild**. Routing: wouter. Server state:
TanStack Query. Client state: zustand. UI: shadcn/ui + Tailwind CSS v4 + clsx.

Brand is **Apex** (`t('brand')`; logo `shared/ui/logo.tsx`; favicon in
`public/`).

**Navigation is backend-driven** — it comes from the nav service (`GET /api/nav`,
`entities/nav`), so never hard-code menu items here. Desktop = side menu
(`widgets/side-nav`, `hidden md:block`); mobile (<md) = Instagram-style bottom
bar (`widgets/bottom-nav`). The **header (`widgets/header`) is minimal**: brand +
user menu only, no nav. Filter items with `visibleNav(items, placement, {isAuthed,
flags})` — the service ships the whole menu + gating metadata and the client
decides. Labels are i18n KEYS (`t(item.labelKey)`); icons are names resolved by
`<NavIcon name={item.icon} />`. Bottom-bar rules: **max 5 buttons**, overflow
folds into a right-side "More"; the **profile/user menu stays in the header on
all viewports** (never in the bar). Keep `main` padded (`pb-24 md:pb-8`) so the
bar never covers content.

Locales: bundled `en`+`ru` in `shared/i18n/locales/*.ts` are the source of
truth; `npm run gen:locales` (part of `npm run build`) exports them + a
manifest to `static/locales/`, served at `/media/locales/*`. The language menu
lists the manifest (`useAvailableLocales`), and `setLanguage` fetches
non-bundled locales from static — so new languages can ship without an app
deploy. Boot requests (`/api/auth/me`, `/api/features`, locale manifest) are
prefetched in `app/index.tsx` + `<link rel="preload">` in `index.html` (SPA
preloading, NOT SSR — SSR would need a framework migration).

The `custom` theme is user-configurable: `shared/theme/model/custom-vars.ts`
(localStorage `custom-theme-vars`, applied as inline CSS vars incl.
`--app-font`) with the UI in `features/customize-theme` on the Profile page.

## Feature-Sliced Design

Layers under `src/`, importing **downward only** (a layer may import from layers
below it, never above or sideways at the same level except via public API):

```
app → pages → widgets → features → entities → shared
```

- `app/` — init: providers (Query, Theme), router, global styles, entry.
- `pages/` — one folder per route; compose widgets/features. Thin. (planner,
  this-week, garage, setups, goals, fuel, …)
- `widgets/` — composite UI blocks (header, bottom-nav, user-menu).
- `features/` — user interactions (auth, fuel-calculator, season-planner,
  manage-content, setups-manager, goal-tracker, customize-theme, link-iracing).
- `entities/` — business entities + their API hooks (viewer, features, iracing,
  driver, planner, setups, goals).
- `shared/` — reusable, domain-agnostic: `ui/` (shadcn — incl. `select`,
  `textarea`), `lib/` (`cn`), `api/` (client + query client), `config/`, `theme/`.

Each slice exposes a public API via its `index.ts`; import from the slice root
(`@/features/auth`), not deep paths.

## Conventions

- **Path alias** `@/` → `src/`.
- **Styling**: only semantic tokens (`bg-background`, `text-foreground`,
  `text-muted-foreground`, `border`, `bg-primary`, `text-destructive`, …). Never
  hard-code hex/rgb — themes (light/dark/custom) work by swapping CSS variables.
- **Responsive**: mobile-first. Must work 320px→1920px. Stack grids on mobile
  (`grid gap-6 md:grid-cols-2`), wrap tables in `overflow-x-auto`, keep the
  header nav behind the mobile menu below `md`.
- **Env**: use `import.meta.env.PUBLIC_*` (never `process.env` — no `process`
  global in the browser; it throws).
- **Data fetching**: queries/mutations via TanStack Query in the entity/feature
  `api/` folder; components read hooks, they don't call `fetch` directly.
- **Forms**: `react-hook-form` + `zod` via `@hookform/resolvers/zod`. Store i18n
  KEYS as the zod messages and render them with `t()` so errors localize (see
  `features/auth`, `setups-manager`, `goal-tracker`). Use the shared Radix
  `Select`/`Textarea` (never native `<select>` or `required`-only validation).
  `Input`/`Textarea` forward refs and style `aria-[invalid=true]`; number fields
  register with `{ valueAsNumber: true }` (avoid `z.coerce`, it breaks RHF types).

## i18n & theming

- Strings go through `useTranslation()` from `@/shared/i18n`; add keys to BOTH
  `locales/en.ts` and `locales/ru.ts` (ru is typed against en, so it must match).
- Language + theme are switched from the avatar/preferences menu (`widgets/user-menu`).

## Storybook & E2E

- Stories: co-locate `*.stories.tsx` next to components. `npm run storybook`.
- E2E (Playwright) in `e2e/*.spec.ts`, run against the live stack on :3000:
  `npx playwright install chromium` once, then `npm run e2e`.

## a11y

Keep: the skip-link in `app/App.tsx`, `aria-label` on icon-only buttons,
`aria-current="page"` on nav links, labeled form fields (`htmlFor`/`id`),
Radix primitives for dialogs/menus (they handle focus + ARIA), and
`cursor-pointer` on interactive elements.

## Checks (run before finishing)

`npm run typecheck && npm run lint && npm test && npm run build`
E2E: `npm run e2e` (stack must be up; Chromium-only — use Pixel device
profiles for mobile specs, iPhone profiles need WebKit which isn't installed).

## Adding a shadcn component

`npx shadcn@latest add <name>` — `components.json` aliases resolve into
`src/shared/ui`. Keep components token-driven so they theme correctly.
