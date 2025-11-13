import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { Inject, Injectable } from '@nestjs/common';
import { Cache } from 'cache-manager';

@Injectable()
export class AppService {
  // constructor(@Inject(CACHE_MANAGER) private cacheManager: Cache) {}

  // async setCache(key: string, value: any) {
  //   await this.cacheManager.set(key, value);
  // }

  // async getCache(key: string) {
  //   return await this.cacheManager.get(key);
  // }

  getHello(): string {
    return 'Hello World!';
  }
}
