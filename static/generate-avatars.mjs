// Generates 8 default avatar SVGs (racing helmet on a gradient) into ./avatars.
// Run: node generate-avatars.mjs
import { writeFileSync, mkdirSync } from 'node:fs';
import { dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const outDir = dirname(fileURLToPath(import.meta.url)) + '/avatars';
mkdirSync(outDir, { recursive: true });

const hues = [8, 45, 90, 150, 190, 220, 275, 320];

function avatar(h) {
  const h2 = (h + 28) % 360;
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128" width="128" height="128" role="img" aria-label="avatar">
  <defs>
    <linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="hsl(${h} 72% 56%)"/>
      <stop offset="1" stop-color="hsl(${h2} 66% 42%)"/>
    </linearGradient>
  </defs>
  <rect width="128" height="128" fill="url(#g)"/>
  <path d="M0 102 L128 46 L128 128 L0 128 Z" fill="#ffffff" opacity="0.10"/>
  <g fill="#ffffff" opacity="0.9">
    <rect x="96" y="10" width="10" height="10"/><rect x="116" y="10" width="10" height="10"/>
    <rect x="106" y="20" width="10" height="10"/>
    <rect x="96" y="30" width="10" height="10"/><rect x="116" y="30" width="10" height="10"/>
  </g>
  <g transform="translate(56 76)">
    <path d="M-32 2 a32 30 0 0 1 64 0 q0 11 -9 13 l-3 1 h-42 q-10 -1 -10 -14 z" fill="#ffffff" opacity="0.95"/>
    <path d="M-24 -3 a24 20 0 0 1 47 5 l-47 0 z" fill="#0b0b0b" opacity="0.55"/>
  </g>
</svg>
`;
}

for (let i = 0; i < hues.length; i++) {
  writeFileSync(`${outDir}/avatar-${i + 1}.svg`, avatar(hues[i]));
}
console.log(`generated ${hues.length} avatars in ${outDir}`);
