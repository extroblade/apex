import { cn } from '@/shared/lib/utils';
import { categoryMeta } from '@/shared/lib/racing-categories';

/**
 * djb2 — tiny deterministic string hash; the same name always renders the same
 * thumbnail, across sessions and devices.
 */
function hash(s: string): number {
  let h = 5381;
  for (let i = 0; i < s.length; i++) h = (h * 33) ^ s.charCodeAt(i);
  return h >>> 0;
}

/**
 * Generated identity art for a catalog item — the app's OWN artwork, so no
 * third-party images are ever shipped. The base gradient keeps the category
 * recognizable (its hue family + icon); the item `name` seeds a deterministic
 * variation — hue shift, gradient angle, and two translucent "livery" stripes —
 * so every car/track/series still looks distinct.
 */
export function CatalogThumbnail({
  category,
  name,
  className,
}: {
  category: string;
  /** Item name; seeds the per-item variation. Omit for a plain category tile. */
  name?: string;
  className?: string;
}) {
  const { Icon, hue } = categoryMeta(category);
  const h = name ? hash(name) : 0;

  // Stay in the category's hue family (±16°) so color still signals category.
  const h1 = (hue + (h % 33) - 16 + 360) % 360;
  const h2 = (h1 + 28 + ((h >> 5) % 20)) % 360;
  const angle = 115 + ((h >> 9) % 50); // 115°–164°
  // Two translucent stripes at hashed offsets — an abstract racing livery.
  const s1 = 22 + ((h >> 13) % 30); // 22–51%
  const s2 = 58 + ((h >> 17) % 26); // 58–83%

  const layers = [
    name
      ? `linear-gradient(${(angle + 62) % 360}deg, transparent ${s1}%, hsl(0 0% 100% / 0.14) ${s1}%, hsl(0 0% 100% / 0.14) ${s1 + 7}%, transparent ${s1 + 7}%)`
      : '',
    name
      ? `linear-gradient(${(angle + 62) % 360}deg, transparent ${s2}%, hsl(0 0% 0% / 0.18) ${s2}%, hsl(0 0% 0% / 0.18) ${s2 + 5}%, transparent ${s2 + 5}%)`
      : '',
    `linear-gradient(${angle}deg, hsl(${h1} 66% 46%), hsl(${h2} 60% 32%))`,
  ]
    .filter(Boolean)
    .join(', ');

  return (
    <div
      className={cn(
        'flex shrink-0 items-center justify-center overflow-hidden rounded-md text-white/90',
        className,
      )}
      style={{ background: layers }}
      aria-hidden
    >
      <Icon className="size-1/2 drop-shadow-sm" />
    </div>
  );
}
