import { HomeService } from './home.service';
import type { ForwardHeaders, UpstreamService } from '../upstream/upstream.service';

// A fake upstream keyed by path, so the reshape logic is tested with no network.
// getApi/getNav resolve from the same map — the test doesn't care which host.
function fakeUpstream(byPath: Record<string, unknown>): UpstreamService {
  const get = async <T>(path: string) => (byPath[path] ?? null) as T | null;
  return { getApi: get, getNav: get, ok: async () => true } as unknown as UpstreamService;
}

const NAV = {
  items: [
    {
      key: 'home',
      labelKey: 'nav.home',
      href: '/',
      icon: 'home',
      placements: ['side', 'bottom'],
      order: 10,
      requiresAuth: false,
    },
    {
      key: 'garage',
      labelKey: 'nav.garage',
      href: '/garage',
      icon: 'warehouse',
      placements: ['side', 'bottom'],
      order: 40,
      requiresAuth: true,
    },
    {
      key: 'drivers',
      labelKey: 'nav.drivers',
      href: '/drivers',
      icon: 'users',
      placements: ['side', 'bottom'],
      order: 70,
      requiresAuth: true,
      featureFlag: 'iracing_oauth',
    },
    {
      key: 'sideonly',
      labelKey: 'nav.sideonly',
      href: '/x',
      icon: 'dot',
      placements: ['side'],
      order: 5,
      requiresAuth: false,
    },
  ],
};

const noHeaders: ForwardHeaders = {};

describe('HomeService', () => {
  it('returns only public, bottom-placed items for an anonymous user', async () => {
    const svc = new HomeService(
      fakeUpstream({
        '/api/auth/me': null,
        '/api/nav': NAV,
        '/api/features': { iracing_oauth: false },
      }),
    );

    const res = await svc.home(noHeaders);

    expect(res.user).toBeNull();
    // home only: garage (auth) + drivers (auth/flag) hidden, sideonly not bottom.
    expect(res.menu.map((m) => m.key)).toEqual(['home']);
    // Menu items are reshaped to the mobile fields only.
    expect(Object.keys(res.menu[0])).toEqual(['key', 'labelKey', 'href', 'icon']);
  });

  it('includes auth items for a signed-in user but still gates by feature flag', async () => {
    const svc = new HomeService(
      fakeUpstream({
        '/api/auth/me': { email: 'a@b.com', nickname: 'Al' },
        '/api/nav': NAV,
        '/api/features': { iracing_oauth: false },
      }),
    );

    const res = await svc.home(noHeaders);

    expect(res.user?.email).toBe('a@b.com');
    expect(res.menu.map((m) => m.key)).toEqual(['home', 'garage']); // drivers still gated
  });

  it('reveals flag-gated items once the flag is on, ordered by `order`', async () => {
    const svc = new HomeService(
      fakeUpstream({
        '/api/auth/me': { email: 'a@b.com' },
        '/api/nav': NAV,
        '/api/features': { iracing_oauth: true },
      }),
    );

    const res = await svc.home(noHeaders);
    expect(res.menu.map((m) => m.key)).toEqual(['home', 'garage', 'drivers']);
  });

  it('degrades to an empty menu when upstream is unavailable', async () => {
    const svc = new HomeService(
      fakeUpstream({ '/api/auth/me': null, '/api/nav': null, '/api/features': null }),
    );

    const res = await svc.home(noHeaders);
    expect(res).toEqual({ user: null, menu: [], features: {} });
  });
});
