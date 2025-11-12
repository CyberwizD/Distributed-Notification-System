import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { TemplatesModule } from './templates/templates.module';
import { HealthModule } from './health/health.module';
import { PrismaService } from './config/prisma.service';

@Module({
  imports: [ConfigModule.forRoot({ isGlobal: true }), TemplatesModule, HealthModule],
  controllers: [AppController],
  providers: [AppService, PrismaService],
  exports: [PrismaService],
})
export class AppModule {}
