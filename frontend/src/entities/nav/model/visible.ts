import type { NavItem } from '../api/use-nav';

export type Placement = 'side' | 'bottom';

export interface NavContext {
  isAuthed: boolean;
  /** The feature-flag map (from GET /api/features). */
  flags: Record<string, boolean>;
}

/**
 * Picks the items a given viewer should see in one placement.
 *
 * The nav service ships the whole menu plus its gating metadata and leaves the
 * decision to the client, which already knows the viewer and the flags. This is
 * safe because navigation isn't a security boundary — the API enforces auth on
 * every route, so hiding a link is purely cosmetic.
 */
export function visibleNav(
  items: NavItem[],
  placement: Placement,
  { isAuthed, flags }: NavContext,
): NavItem[] {
  return items
    .filter((i) => i.placements.includes(placement))
    .filter((i) => !i.requiresAuth || isAuthed)
    .filter((i) => !i.featureFlag || flags[i.featureFlag] === true)
    .sort((a, b) => a.order - b.order);
}
