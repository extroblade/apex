import { Controller, Get, Req } from '@nestjs/common';
import type { Request } from 'express';

import { HomeService, MobileHome } from './home.service';

@Controller('bff')
export class HomeController {
  constructor(private readonly home: HomeService) {}

  // Mobile home screen in one call. Forwards the caller's auth to the Go API.
  @Get('home')
  getHome(@Req() req: Request): Promise<MobileHome> {
    return this.home.home({
      cookie: req.headers.cookie,
      authorization: req.headers.authorization,
    });
  }
}
