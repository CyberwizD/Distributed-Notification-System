import { Injectable } from '@nestjs/common';
import {
  HealthIndicatorResult,
  HealthIndicatorService,
} from '@nestjs/terminus';
import { RedisService } from '../config/redis.config';

@Injectable()
export class RedisHealthIndicator {
  constructor(
    private readonly redisService: RedisService,
    private readonly health: HealthIndicatorService,
  ) {}

  async isHealthy(key: string): Promise<HealthIndicatorResult> {
    const client = this.redisService.getClient();
    if (!client || client.status !== 'ready') {
      return this.health.check(key).down('Redis is not connected');
    }
    return this.health.check(key).up();
  }
}
