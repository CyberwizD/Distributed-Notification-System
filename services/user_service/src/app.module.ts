import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { AuthModule } from './auth/auth.module';
import { UsersModule } from './users/users.module';

import { ConfigModule, ConfigService } from '@nestjs/config';
import { CacheModule } from '@nestjs/cache-manager';
import { redisStore } from 'cache-manager-redis-yet';
import { HealthModule } from './health/health.module';
import { PrismaModule } from './prisma/prisma.module';
import { RabbitMQModule } from './common/rabbitmq/rabbitmq.module';
import databaseConfig from './config/database.config';
import redisConfig from './config/redis.config';
import jwtConfig from './config/jwt.config';
import serviceConfig from './config/service.config';


@Module({
  imports: [
    ConfigModule.forRoot({ isGlobal: true, load: [databaseConfig, redisConfig, jwtConfig, serviceConfig] }),
    CacheModule.registerAsync({
      isGlobal: true,
      imports: [ConfigModule],
      useFactory: async (configService: ConfigService) => ({
        store: await redisStore({
          socket: {
            host: configService.get('redis.url') || 'localhost',
            port: 6379,
          },
        }),
        ttl: configService.get('redis.ttl') * 1000, // Convert to milliseconds
      }),
      inject: [ConfigService],
    }),
    AuthModule,
    UsersModule,
    HealthModule,
    RabbitMQModule,
    PrismaModule,
  ],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule { }
