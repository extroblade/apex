import { Controller, Get } from '@nestjs/common';

import { UpstreamService } from '../upstream/upstream.service';

@Controller('bff')
export class HealthController {
  constructor(private readonly upstream: UpstreamService) {}

  @Get('health')
  async health(): Promise<{ ok: boolean; upstream: boolean }> {
    return { ok: true, upstream: await this.upstream.ok('/api/health') };
  }
}
