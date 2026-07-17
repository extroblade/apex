import { Module } from '@nestjs/common';

import { HealthController } from './health/health.controller';
import { HomeController } from './home/home.controller';
import { HomeService } from './home/home.service';
import { MetricsController } from './metrics/metrics.controller';
import { UpstreamService } from './upstream/upstream.service';

@Module({
  controllers: [HomeController, HealthController, MetricsController],
  providers: [UpstreamService, HomeService],
})
export class AppModule {}
