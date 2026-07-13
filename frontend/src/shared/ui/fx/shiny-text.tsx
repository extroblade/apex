import { cn } from '@/shared/lib/utils';

/** Text with a slow shimmer sweep (reactbits-style "Shiny Text"). */
export function ShinyText({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return <span className={cn('fx-shiny-text', className)}>{children}</span>;
}
