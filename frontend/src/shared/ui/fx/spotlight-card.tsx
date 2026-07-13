import { useRef } from 'react';

import { cn } from '@/shared/lib/utils';

/**
 * A card whose border glow follows the pointer (reactbits-style "Spotlight
 * Card"). Pure CSS highlight driven by --fx-x/--fx-y custom properties.
 */
export function SpotlightCard({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  const ref = useRef<HTMLDivElement>(null);

  return (
    <div
      ref={ref}
      onMouseMove={(e) => {
        const el = ref.current;
        if (!el) return;
        const rect = el.getBoundingClientRect();
        el.style.setProperty('--fx-x', `${e.clientX - rect.left}px`);
        el.style.setProperty('--fx-y', `${e.clientY - rect.top}px`);
      }}
      className={cn(
        // Mirrors shared/ui Card so CardHeader/CardContent compose inside it.
        'fx-spotlight relative flex flex-col gap-6 rounded-xl border bg-card py-6 text-card-foreground shadow-sm',
        className,
      )}
    >
      {children}
    </div>
  );
}
