import { Counter, Histogram, Registry, collectDefaultMetrics } from 'prom-client';

// One registry for the whole service. Default process/node metrics + our own.
export const registry = new Registry();
collectDefaultMetrics({ register: registry });

export const httpRequests = new Counter({
  name: 'bff_http_requests_total',
  help: 'HTTP requests by method, path and status.',
  labelNames: ['method', 'path', 'status'],
  registers: [registry],
});

export const httpDuration = new Histogram({
  name: 'bff_http_request_duration_seconds',
  help: 'HTTP request duration in seconds by method and path.',
  labelNames: ['method', 'path'],
  registers: [registry],
});

const domainCounters = new Map<string, Counter<string>>();

/**
 * Generic domain counter, mirroring the frontend's counter('event')(labels) and
 * the Go backend's metrics.Count. Lazily creates the collector on first use; the
 * label keys fix its label set, so call a given name with the same keys.
 */
export function count(name: string, help: string, labels: Record<string, string>): void {
  let c = domainCounters.get(name);
  if (!c) {
    c = new Counter({
      name,
      help,
      labelNames: Object.keys(labels),
      registers: [registry],
    });
    domainCounters.set(name, c);
  }
  c.inc(labels);
}
