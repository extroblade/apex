import { Injectable } from '@nestjs/common';

import { count } from '../metrics/metrics';
import { ForwardHeaders, UpstreamService } from '../upstream/upstream.service';

interface NavItem {
  key: string;
  labelKey: string;
  href: string;
  icon: string;
  placements: string[];
  order: number;
  requiresAuth: boolean;
  featureFlag?: string;
}

interface Viewer {
  email: string;
  nickname?: string;
  avatarUrl?: string;
}

/** A ready-to-render mobile menu entry (gating already resolved server-side). */
export interface MobileMenuItem {
  key: string;
  labelKey: string;
  href: string;
  icon: string;
}

export interface MobileHome {
  user: Viewer | null;
  menu: MobileMenuItem[];
  features: Record<string, boolean>;
}

/**
 * Aggregates the three calls a mobile home screen needs (session, menu, flags)
 * into ONE response, and does the nav gating here so the app doesn't reimplement
 * visibleNav — the classic BFF win: fewer round trips, thinner client.
 */
@Injectable()
export class HomeService {
  constructor(private readonly upstream: UpstreamService) {}

  async home(headers: ForwardHeaders): Promise<MobileHome> {
    const [me, nav, features] = await Promise.all([
      this.upstream.getApi<Viewer>('/api/auth/me', headers),
      this.upstream.getNav<{ items: NavItem[] }>('/api/nav', headers),
      this.upstream.getApi<Record<string, boolean>>('/api/features', headers),
    ]);

    const flags = features ?? {};
    const isAuthed = me != null;

    const menu: MobileMenuItem[] = (nav?.items ?? [])
      .filter((i) => i.placements.includes('bottom'))
      .filter((i) => !i.requiresAuth || isAuthed)
      .filter((i) => !i.featureFlag || flags[i.featureFlag] === true)
      .sort((a, b) => a.order - b.order)
      .map((i) => ({ key: i.key, labelKey: i.labelKey, href: i.href, icon: i.icon }));

    count('bff_home_requests_total', 'Mobile home aggregations, by auth state.', {
      authed: String(isAuthed),
    });

    return { user: me, menu, features: flags };
  }
}
