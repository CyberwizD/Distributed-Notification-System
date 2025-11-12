import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { TemplatesModule } from './templates/templates.module';
import { HealthModule } from './health/health.module';
import { PrismaService } from './config/prisma.service';

@Module({
  imports: [TemplatesModule, HealthModule],
  controllers: [AppController],
  providers: [AppService, PrismaService],
  exports: [PrismaService],
})
export class AppModule {}
