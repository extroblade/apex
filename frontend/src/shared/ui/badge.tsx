import { cn } from '@/shared/lib/utils';

/** A small pill for labels like a category tag. */
function Badge({ className, ...props }: React.ComponentProps<'span'>) {
  return (
    <span
      data-slot="badge"
      className={cn(
        'inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium text-muted-foreground',
        className,
      )}
      {...props}
    />
  );
}

export { Badge };
