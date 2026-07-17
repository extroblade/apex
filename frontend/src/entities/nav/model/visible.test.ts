import { describe, it, expect } from 'vitest';

import { visibleNav } from './visible';
import type { NavItem } from '../api/use-nav';

const item = (over: Partial<NavItem> & Pick<NavItem, 'key'>): NavItem => ({
  labelKey: `nav.${over.key}`,
  href: `/${over.key}`,
  icon: 'home',
  placements: ['side', 'bottom'],
  order: 0,
  requiresAuth: false,
  ...over,
});

const ITEMS: NavItem[] = [
  item({ key: 'home', order: 10 }),
  item({ key: 'planner', order: 30, requiresAuth: true }),
  item({ key: 'drivers', order: 20, requiresAuth: true, featureFlag: 'iracing_oauth' }),
  item({ key: 'thisWeek', order: 40, placements: ['side'] }),
];

describe('visibleNav', () => {
  it('hides auth-only items from anonymous viewers', () => {
    const got = visibleNav(ITEMS, 'side', { isAuthed: false, flags: {} });
    expect(got.map((i) => i.key)).toEqual(['home', 'thisWeek']);
  });

  it('hides flag-gated items while the flag is off', () => {
    const got = visibleNav(ITEMS, 'side', {
      isAuthed: true,
      flags: { iracing_oauth: false },
    });
    expect(got.map((i) => i.key)).not.toContain('drivers');
  });

  it('shows flag-gated items once the flag is on, ordered by `order`', () => {
    const got = visibleNav(ITEMS, 'side', {
      isAuthed: true,
      flags: { iracing_oauth: true },
    });
    expect(got.map((i) => i.key)).toEqual(['home', 'drivers', 'planner', 'thisWeek']);
  });

  it('filters by placement', () => {
    const got = visibleNav(ITEMS, 'bottom', { isAuthed: true, flags: {} });
    expect(got.map((i) => i.key)).toEqual(['home', 'planner']);
  });

  it('treats a missing flag as off', () => {
    const got = visibleNav(ITEMS, 'side', { isAuthed: true, flags: {} });
    expect(got.map((i) => i.key)).not.toContain('drivers');
  });
});
