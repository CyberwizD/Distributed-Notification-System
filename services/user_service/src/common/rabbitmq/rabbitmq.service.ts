import { Injectable, OnModuleInit, OnModuleDestroy, Logger } from '@nestjs/common';
import * as amqp from 'amqplib';

@Injectable()
export class RabbitMQService implements OnModuleInit, OnModuleDestroy {
    private readonly logger = new Logger(RabbitMQService.name);
    private connection: amqp.Connection | null = null;
    private channel: amqp.Channel | null = null;
    private isConnected = false;

    async onModuleInit() {
        await this.connect();
    }

    async onModuleDestroy() {
        await this.disconnect();
    }

    private async connect() {
        try {
            const rabbitmqUrl = process.env.RABBITMQ_URL || 'amqp://localhost:5672';

            const connection = await amqp.connect(rabbitmqUrl);
            this.connection = connection;
            this.channel = await connection.createChannel();

            // Declare the main exchange for user events
            await this.channel.assertExchange('user.events', 'topic', {
                durable: true,
                autoDelete: false,
            });

            this.isConnected = true;
            this.logger.log('Successfully connected to RabbitMQ');

            // Handle connection errors
            connection.on('error', (error) => {
                this.logger.error('RabbitMQ connection error:', error);
                this.isConnected = false;
            });

            connection.on('close', () => {
                this.logger.warn('RabbitMQ connection closed');
                this.isConnected = false;
            });

        } catch (error) {
            this.logger.error('Failed to connect to RabbitMQ:', error);
            this.isConnected = false;
        }
    }

    private async disconnect() {
        try {
            if (this.channel) {
                await this.channel.close();
            }
            if (this.connection) {
                await this.connection.close();
            }
            this.logger.log('Disconnected from RabbitMQ');
        } catch (error) {
            this.logger.error('Error disconnecting from RabbitMQ:', error);
        }
    }

    async publish(routingKey: string, message: any) {
        if (!this.isConnected || !this.channel) {
            this.logger.warn(`RabbitMQ not connected. Cannot publish message: ${routingKey}`);
            return;
        }

        try {
            const payload = Buffer.from(JSON.stringify({
                ...message,
                timestamp: new Date().toISOString(),
                source: 'user-service',
                version: '1.0.0'
            }));

            const published = this.channel.publish('user.events', routingKey, payload, {
                persistent: true,
                contentType: 'application/json',
            });

            if (published) {
                this.logger.log(`Event published: ${routingKey}`);
            } else {
                this.logger.warn(`Failed to publish event: ${routingKey}`);
            }

            return published;
        } catch (error) {
            this.logger.error(`Error publishing event ${routingKey}:`, error);
            throw error;
        }
    }

    async isHealthy(): Promise<boolean> {
        return this.isConnected && this.channel !== undefined;
    }
}