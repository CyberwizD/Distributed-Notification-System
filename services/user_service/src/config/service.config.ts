import { registerAs } from '@nestjs/config';

export default registerAs('service', () => ({
  name: 'user-service',
  version: '1.0.0',
  port: parseInt(process.env.PORT ?? '3001', 10) || 3001,
  internalApiKey: process.env.INTERNAL_API_KEY || 'default-internal-key-change-in-production',
  rabbitMQ: {
    url: process.env.RABBITMQ_URL || 'amqp://localhost:5672',
    exchanges: {
      userEvents: 'user.events'
    }
  }
}));