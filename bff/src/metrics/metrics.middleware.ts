import type { NextFunction, Request, Response } from 'express';

import { httpDuration, httpRequests } from './metrics';

/**
 * Records request count + duration on finish. The BFF has a small, fixed set of
 * routes, so labeling by req.path keeps cardinality bounded without needing the
 * matched route pattern.
 */
export function metricsMiddleware(req: Request, res: Response, next: NextFunction): void {
  const stop = httpDuration.startTimer({ method: req.method, path: req.path });
  res.on('finish', () => {
    stop();
    httpRequests.inc({
      method: req.method,
      path: req.path,
      status: String(res.statusCode),
    });
  });
  next();
}
