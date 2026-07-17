import { Test } from '@nestjs/testing';
import type { INestApplication } from '@nestjs/common';
import request from 'supertest';

import { HealthController } from './health.controller';
import { UpstreamService } from '../upstream/upstream.service';

describe('Health (e2e)', () => {
  let app: INestApplication;

  beforeAll(async () => {
    const moduleRef = await Test.createTestingModule({
      controllers: [HealthController],
      providers: [{ provide: UpstreamService, useValue: { ok: async () => false } }],
    }).compile();

    app = moduleRef.createNestApplication();
    await app.init();
  });

  afterAll(async () => {
    await app.close();
  });

  it('GET /bff/health reports service up and upstream state', async () => {
    const res = await request(app.getHttpServer()).get('/bff/health').expect(200);
    expect(res.body).toEqual({ ok: true, upstream: false });
  });
});
