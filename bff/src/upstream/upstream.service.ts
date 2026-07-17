import { Injectable } from '@nestjs/common';

/** Headers forwarded from the caller so upstreams see the mobile user's auth. */
export interface ForwardHeaders {
  cookie?: string;
  authorization?: string;
}

// The BFF is the aggregation layer, so it knows the service topology: the Go API
// and the nav service are distinct hosts (nginx splits them for the web app; the
// BFF talks to each directly).
const API_BASE = process.env.API_BASE_URL ?? 'http://backend:8080';
const NAV_BASE = process.env.NAV_BASE_URL ?? 'http://nav:8081';

/**
 * Thin client for the upstream services. The BFF owns no auth — it forwards the
 * caller's session cookie / bearer token so upstream authorization is unchanged.
 * Any failure returns null so callers degrade gracefully.
 */
@Injectable()
export class UpstreamService {
  /** GET JSON from the Go API. */
  getApi<T>(path: string, headers: ForwardHeaders = {}): Promise<T | null> {
    return this.fetchJson<T>(API_BASE, path, headers);
  }

  /** GET JSON from the nav service. */
  getNav<T>(path: string, headers: ForwardHeaders = {}): Promise<T | null> {
    return this.fetchJson<T>(NAV_BASE, path, headers);
  }

  /** Liveness probe against the Go API. */
  async ok(path: string): Promise<boolean> {
    try {
      const res = await fetch(`${API_BASE}${path}`);
      return res.ok;
    } catch {
      return false;
    }
  }

  private async fetchJson<T>(
    base: string,
    path: string,
    headers: ForwardHeaders,
  ): Promise<T | null> {
    const fwd: Record<string, string> = {};
    if (headers.cookie) fwd.cookie = headers.cookie;
    if (headers.authorization) fwd.authorization = headers.authorization;

    try {
      const res = await fetch(`${base}${path}`, { headers: fwd });
      if (!res.ok) return null;
      return (await res.json()) as T;
    } catch {
      return null;
    }
  }
}
