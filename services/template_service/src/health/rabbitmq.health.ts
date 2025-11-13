import { Injectable } from '@nestjs/common';
import {
  HealthIndicatorResult,
  HealthIndicatorService,
} from '@nestjs/terminus';
import { RabbitMQService } from '../config/rabbitmq.config';

@Injectable()
export class RabbitMQHealthIndicator {
  constructor(
    private readonly rabbitMQService: RabbitMQService,
    private readonly health: HealthIndicatorService,
  ) {}

  async isHealthy(key: string): Promise<HealthIndicatorResult> {
    return this.rabbitMQService.getChannel()
      ? this.health.check(key).up()
      : this.health.check(key).down('RabbitMQ is not connected');
  }
}
