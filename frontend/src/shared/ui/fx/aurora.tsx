import { cn } from '@/shared/lib/utils';

/**
 * Soft animated gradient blobs behind a section (reactbits-style "Aurora").
 * Colors come from theme tokens, so light/dark/custom themes tint it — and it
 * stays decorative: pointer-events-none, aria-hidden, reduced-motion aware.
 *
 * The parent must be `relative`; content should sit in a `relative` sibling so
 * it stacks above.
 */
export function Aurora({ className }: { className?: string }) {
  return (
    <div
      aria-hidden
      className={cn('pointer-events-none absolute inset-0 overflow-hidden', className)}
    >
      <div className="fx-aurora-blob absolute -left-24 -top-24 size-96 rounded-full bg-primary/20 blur-3xl" />
      <div
        className="fx-aurora-blob absolute -right-16 top-1/4 size-80 rounded-full bg-ring/25 blur-3xl"
        style={{ animationDelay: '-6s', animationDuration: '20s' }}
      />
      <div
        className="fx-aurora-blob absolute -bottom-24 left-1/3 size-72 rounded-full bg-primary/10 blur-3xl"
        style={{ animationDelay: '-11s', animationDuration: '24s' }}
      />
    </div>
  );
}
