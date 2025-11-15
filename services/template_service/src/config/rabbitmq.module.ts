
import { Module } from '@nestjs/common';
import { RabbitMQService } from './rabbitmq.config';

@Module({
  providers: [RabbitMQService],
  exports: [RabbitMQService],
})
export class RabbitMQModule {}
