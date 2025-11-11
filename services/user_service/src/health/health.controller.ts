import { Controller, Get, Inject } from '@nestjs/common';
import { ApiTags, ApiOperation } from '@nestjs/swagger';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import type { Cache } from 'cache-manager';
import { PrismaService } from '../prisma/prisma.service';

@ApiTags('Health')
@Controller('health')
export class HealthController {
    constructor(
        private prisma: PrismaService,
        @Inject(CACHE_MANAGER) private cacheManager: Cache,
    ) { }

    @Get()
    @ApiOperation({ summary: 'Health check' })
    async check() {
        const dbStatus = await this.checkDatabase();
        const cacheStatus = await this.checkCache();

        const status = dbStatus && cacheStatus ? 'healthy' : 'unhealthy';

        return {
            status,
            timestamp: new Date().toISOString(),
            checks: {
                database: dbStatus ? 'connected' : 'disconnected',
                cache: cacheStatus ? 'connected' : 'disconnected',
            },
        };
    }

    private async checkDatabase(): Promise<boolean> {
        try {
            await this.prisma.$queryRaw`SELECT 1`;
            return true;
        } catch {
            return false;
        }
    }

    private async checkCache(): Promise<boolean> {
        try {
            const testKey = 'health-check';
            await this.cacheManager.set(testKey, 'ok', 1000);
            const value = await this.cacheManager.get(testKey);
            await this.cacheManager.del(testKey);
            return value === 'ok';
        } catch {
            return false;
        }
    }
}