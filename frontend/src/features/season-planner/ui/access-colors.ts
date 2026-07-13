import type { TrackAccess } from '@/entities/planner';

/**
 * Cell background/ring classes for a schedule cell. A planned race wins over the
 * access color with a warm amber highlight so "I'm racing this" reads instantly
 * and distinctly from the green/aquamarine/red availability states.
 */
export function accessCellClasses(access: TrackAccess, planned: boolean): string {
  if (planned) {
    return 'bg-amber-400/25 text-foreground ring-2 ring-amber-500 hover:ring-amber-500 font-medium';
  }
  switch (access) {
    case 'free':
      return 'bg-green-500/20 text-foreground ring-1 ring-green-600/70 hover:ring-green-600';
    case 'owned':
      return 'bg-teal-500/20 text-foreground ring-1 ring-teal-600/70 hover:ring-teal-600';
    default:
      return 'bg-red-500/15 text-muted-foreground ring-1 ring-red-600/50 hover:ring-red-600';
  }
}

/** Text color matching the access state, for inline track names. */
export function accessTextClass(access: TrackAccess): string {
  switch (access) {
    case 'free':
      return 'text-green-600';
    case 'owned':
      return 'text-teal-600';
    default:
      return 'text-red-500';
  }
}
