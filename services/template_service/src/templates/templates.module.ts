import { Module } from '@nestjs/common';
import { TemplatesService } from './templates.service';
import { TemplatesController } from './templates.controller';
// import { PrismaService } from '../config/database.config';
import { SubstitutionService } from '../substitution/substitution.service';
import { PrismaService } from 'src/config/prisma.service';

@Module({
  controllers: [TemplatesController],
  providers: [TemplatesService, PrismaService, SubstitutionService],
})
export class TemplatesModule {}
