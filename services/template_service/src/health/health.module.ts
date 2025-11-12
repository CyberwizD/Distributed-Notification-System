import { Module } from '@nestjs/common';
import { TerminusModule } from '@nestjs/terminus';
import { HealthController } from './health.controller';
import { PrismaHealthIndicator } from './prisma.health';
import { PrismaService } from 'src/config/prisma.service';
import { RedisHealthIndicator } from './redis.health';
import { RabbitMQHealthIndicator } from './rabbitmq.health';
import { RedisService } from 'src/config/redis.config';
import { RabbitMQService } from 'src/config/rabbitmq.config';

@Module({
  imports: [TerminusModule],
  controllers: [HealthController],
  providers: [
    PrismaHealthIndicator,
    PrismaService,
    // RedisHealthIndicator,
    // RedisService,
    RabbitMQHealthIndicator,
    RabbitMQService,
  ],
})
export class HealthModule {}
