// Exports the authored locale bundles (en, ru) to the backend, which embeds and
// seeds them into the `locales` table and serves them at /api/locales/{code}.
// en.ts stays the source of truth AND the Translation type; ru.ts is typed
// against it. Regenerate after editing either, and commit the JSON so the
// backend build can go:embed it.
//
// The backend owns the LIST of languages now (DB-driven), so adding a language
// no longer means a frontend build — insert a row (or add a bundle here for a
// built-in). No manifest is emitted.
//
// Run: node scripts/gen-locales.mjs   (also wired as `npm run gen:locales`)
// Requires Node 22.6+ (native TypeScript type-stripping for the .ts imports).
import { writeFileSync, mkdirSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));

const { en } = await import(resolve(here, '../src/shared/i18n/locales/en.ts'));
const { ru } = await import(resolve(here, '../src/shared/i18n/locales/ru.ts'));

// Backend embed dir (go:embed data/*.json). During the frontend Docker build
// this path resolves to a throwaway location inside the container — the frontend
// image doesn't use it; only the committed copy feeds the backend build.
const out = resolve(here, '../../backend/internal/locales/data');
mkdirSync(out, { recursive: true });
writeFileSync(`${out}/en.json`, JSON.stringify(en, null, 2) + '\n');
writeFileSync(`${out}/ru.json`, JSON.stringify(ru, null, 2) + '\n');
console.log(`wrote en, ru bundles to ${out}`);
