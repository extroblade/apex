import { cn } from '@/shared/lib/utils';

/**
 * The Apex mark: a racing line arcing through an apex point. Uses currentColor
 * so it inherits the theme (wrap in text-primary for the brand color).
 */
export function AppLogo({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 32 32"
      fill="none"
      className={cn('size-6', className)}
      role="img"
      aria-label="Apex"
    >
      <path
        d="M5 27 C 5 14, 14 5, 27 5"
        stroke="currentColor"
        strokeWidth="3.2"
        strokeLinecap="round"
      />
      <circle cx="11.4" cy="11.4" r="2.9" fill="currentColor" />
    </svg>
  );
}
