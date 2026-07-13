import { cn } from '@/shared/lib/utils';
import { categoryMeta } from '@/shared/lib/racing-categories';

/**
 * A generated "image" for a catalog item: a category-tinted gradient with the
 * category icon. Real iRacing content art needs the authenticated CDN, so this
 * gives every car/track/series a consistent, recognizable thumbnail.
 */
export function CatalogThumbnail({
  category,
  className,
}: {
  category: string;
  className?: string;
}) {
  const { Icon, hue } = categoryMeta(category);
  return (
    <div
      className={cn(
        'flex shrink-0 items-center justify-center rounded-md text-white/90',
        className,
      )}
      style={{
        background: `linear-gradient(135deg, hsl(${hue} 66% 46%), hsl(${(hue + 32) % 360} 60% 34%))`,
      }}
      aria-hidden
    >
      <Icon className="size-1/2" />
    </div>
  );
}
