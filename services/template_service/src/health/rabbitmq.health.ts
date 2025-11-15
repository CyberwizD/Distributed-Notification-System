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
    const channel = this.rabbitMQService.getChannel();
    if (channel) {
      // The library keeps the connection object available.
      // If 'closeReason' is present, the connection is closed or closing.
      if ((channel.connection as any).closeReason) {
        return this.health.check(key).down('RabbitMQ connection is closed.');
      }
      return this.health.check(key).up();
    }
    return this.health.check(key).down('RabbitMQ is not connected');
  }
}
