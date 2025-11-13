import { Injectable, OnModuleInit, OnModuleDestroy } from '@nestjs/common';
import { PrismaClient } from '@prisma/client';

@Injectable()
export class PrismaService
  extends PrismaClient
  implements OnModuleInit, OnModuleDestroy
{
  async onModuleInit() {
    try {
      await this.$connect();
      console.log('✅ Prisma connected successfully');
    } catch (error) {
      console.error('❌ Prisma connection failed:', error.message);
      // Depending on your error handling strategy, you might want to process.exit() here
    }
  }

  async onModuleDestroy() {
    await this.$disconnect();
  }
}
