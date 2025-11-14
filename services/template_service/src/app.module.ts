import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { TemplatesModule } from './templates/templates.module';
import { HealthModule } from './health/health.module';
import { PrismaService } from './config/prisma.service';
import { CacheModule } from '@nestjs/cache-manager';
import { max } from 'class-validator';

@Module({
  imports: [
    TemplatesModule,
    HealthModule,
    CacheModule.register({
      max: 1000,
      ttl: 0,
    }),
  ],
  controllers: [AppController],
  providers: [AppService, PrismaService],
  exports: [PrismaService],
})
export class AppModule {}