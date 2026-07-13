// Exports the app locales + a manifest to the static service so translations
// are served (and cached) from /media/locales/*. The manifest drives the
// language menu — drop a new locale JSON + manifest entry to add a language
// without redeploying the app bundle.
//
// Run: node scripts/gen-locales.mjs   (also wired as `npm run gen:locales`)
// Requires Node 22.6+ (native TypeScript type-stripping for the .ts imports).
import { writeFileSync, mkdirSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));

const { en } = await import(resolve(here, '../src/shared/i18n/locales/en.ts'));
const { ru } = await import(resolve(here, '../src/shared/i18n/locales/ru.ts'));

const manifest = {
  locales: [
    { code: 'en', name: 'English' },
    { code: 'ru', name: 'Русский' },
  ],
};

const out = resolve(here, '../../static/locales');
mkdirSync(out, { recursive: true });
writeFileSync(`${out}/en.json`, JSON.stringify(en, null, 2));
writeFileSync(`${out}/ru.json`, JSON.stringify(ru, null, 2));
writeFileSync(`${out}/index.json`, JSON.stringify(manifest, null, 2));
console.log(`wrote ${manifest.locales.length} locales + manifest to ${out}`);
