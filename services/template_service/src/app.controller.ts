import { Controller, Get, Post, Query } from '@nestjs/common';
import { AppService } from './app.service';

@Controller()
export class AppController {
  constructor(private readonly appService: AppService) {}

  // @Post()
  // async setCacheKey(@Query('key') key: string, @Query('value') value: string) {
  //   await this.appService.setCache(key, value);
  //   return { message: 'Cache set successfully', status: 201, success: true };
  // }

  // @Get('/get/:key')
  // async getCacheKey(@Query('key') key: string) {
  //   const value = await this.appService.getCache(key);
  //   return {
  //     message: 'Cache retrieved successfully',
  //     status: 200,
  //     success: true,
  //     value,
  //   };
  // }

  @Get()
  getHello(): string {
    return this.appService.getHello();
  }
}
