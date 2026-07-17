import 'reflect-metadata';
import { NestFactory } from '@nestjs/core';

import { AppModule } from './app.module';
import { metricsMiddleware } from './metrics/metrics.middleware';

async function bootstrap(): Promise<void> {
  const app = await NestFactory.create(AppModule);
  app.use(metricsMiddleware);
  const port = Number(process.env.PORT ?? 8083);
  await app.listen(port, '0.0.0.0');
  // eslint-disable-next-line no-console
  console.log(`bff: listening on :${port}`);
}

void bootstrap();
