import { Controller, Get, Header } from '@nestjs/common';

import { registry } from './metrics';

@Controller()
export class MetricsController {
  // Prometheus exposition. Not proxied publicly — scraped inside apex-net.
  @Get('metrics')
  @Header('Content-Type', 'text/plain; version=0.0.4; charset=utf-8')
  metrics(): Promise<string> {
    return registry.metrics();
  }
}
