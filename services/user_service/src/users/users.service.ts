import { Injectable, NotFoundException, ConflictException, Inject, BadRequestException } from '@nestjs/common';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import type { Cache } from 'cache-manager';
import { PrismaService } from '../prisma/prisma.service';
import { RabbitMQService } from '../common/rabbitmq/rabbitmq.service';
import { CreateUserDto } from './dto/create-user.dto';
import { UpdateUserDto } from './dto/update-user.dto';
import { UpdatePreferenceDto } from './dto/update-preference.dto';

@Injectable()
export class UsersService {
  constructor(
    private prisma: PrismaService,
    @Inject(CACHE_MANAGER) private cacheManager: Cache,
    private rabbitMQService: RabbitMQService,
  ) { }

  async create(createUserDto: CreateUserDto) {
    const { email, password, name, emailEnabled, pushEnabled } = createUserDto;

    // Check if user exists
    const existingUser = await this.prisma.user.findUnique({
      where: { email },
    });

    if (existingUser) {
      throw new ConflictException('User with this email already exists');
    }

    // Create user with preferences
    return this.prisma.user.create({
      data: {
        email,
        password, // Note: In real implementation, hash this password
        name,
        preferences: {
          create: {
            emailEnabled: emailEnabled ?? true,
            pushEnabled: pushEnabled ?? true,
          },
        },
      },
      include: {
        preferences: true,
      },
    });
  }

  async findAll(page: number = 1, limit: number = 10) {
    const skip = (page - 1) * limit;

    const [users, total] = await Promise.all([
      this.prisma.user.findMany({
        skip,
        take: limit,
        include: {
          preferences: true,
        },
        orderBy: { createdAt: 'desc' },
      }),
      this.prisma.user.count(),
    ]);

    return {
      data: users,
      meta: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    };
  }

  async findOne(id: string) {
    const cacheKey = `user:${id}`;
    const cachedUser = await this.cacheManager.get(cacheKey);

    if (cachedUser) {
      return cachedUser;
    }

    const user = await this.prisma.user.findUnique({
      where: { id },
      include: {
        preferences: true,
        deviceTokens: {
          where: { isActive: true },
        },
      },
    });

    if (!user) {
      throw new NotFoundException('User not found');
    }

    // Cache for 5 minutes
    await this.cacheManager.set(cacheKey, user, 300000);

    return user;
  }

  async findByEmail(email: string) {
    return this.prisma.user.findUnique({
      where: { email },
      include: { preferences: true },
    });
  }

  async update(id: string, updateUserDto: UpdateUserDto) {
    await this.findOne(id); // Check if user exists

    const user = await this.prisma.user.update({
      where: { id },
      data: updateUserDto,
      include: { preferences: true },
    });

    // Clear cache
    await this.cacheManager.del(`user:${id}`);

    return user;
  }

  async remove(id: string) {
    await this.findOne(id); // Check if user exists

    await this.prisma.user.update({
      where: { id },
      data: { isActive: false },
    });

    // Clear cache
    await this.cacheManager.del(`user:${id}`);
  }

  async updatePreferences(userId: string, updatePreferenceDto: UpdatePreferenceDto) {
    await this.findOne(userId); // Check if user exists

    const oldPreferences = await this.getPreferences(userId);

    const preferences = await this.prisma.userPreference.upsert({
      where: { userId },
      update: updatePreferenceDto,
      create: {
        userId,
        ...updatePreferenceDto,
      },
    });

    // Publish preferences updated event
    try {
      await this.rabbitMQService.publish('user.preferences.updated', {
        userId,
        oldPreferences,
        newPreferences: preferences,
        changes: Object.keys(updatePreferenceDto)
      });
      console.log('✅ User preferences updated event published');
    } catch (eventError) {
      console.error('❌ Failed to publish preferences updated event:', eventError);
    }

    // Clear cache
    await this.cacheManager.del(`user:${userId}`);

    return preferences;
  }

  async getPreferences(userId: string) {
    const preferences = await this.prisma.userPreference.findUnique({
      where: { userId },
    });

    if (!preferences) {
      // Create default preferences if they don't exist
      return this.prisma.userPreference.create({
        data: { userId },
      });
    }

    return preferences;
  }

  // Methods for other services
  async canReceiveEmail(userId: string): Promise<boolean> {
    const preferences = await this.getPreferences(userId);
    return preferences.emailEnabled;
  }

  async canReceivePush(userId: string): Promise<boolean> {
    const preferences = await this.getPreferences(userId);
    return preferences.pushEnabled;
  }

  async getEmailAddress(userId: string): Promise<string> {
    const user = await this.prisma.user.findUnique({
      where: { id: userId },
      select: { email: true },
    });

    if (!user) {
      throw new NotFoundException('User not found');
    }

    return user.email;
  }

  async getUserContactInfo(userId: string) {
    const user = await this.prisma.user.findUnique({
      where: { id: userId },
      include: {
        preferences: true,
        deviceTokens: {
          where: { isActive: true },
        },
      },
    });

    if (!user) {
      throw new NotFoundException('User not found');
    }

    return {
      email: user.email,
      preferences: user.preferences,
      deviceTokens: user.deviceTokens,
    };
  }

  // Device token management
  async addDeviceToken(userId: string, token: string, platform: string) {
    await this.findOne(userId); // Check if user exists

 if (!token || !platform) {
      throw new BadRequestException('Token and platform are required');
    }

    if (!['ios', 'android', 'web'].includes(platform)) {
      throw new BadRequestException('Platform must be ios, android, or web');
    }

    const deviceToken = await this.prisma.deviceToken.upsert({
      where: { token },
      update: {
        userId,
        platform,
        isActive: true,
      },
      create: {
        userId,
        token,
        platform,
      },
    });

     // Publish device token added event
    try {
      await this.rabbitMQService.publish('user.device-token.added', {
        userId,
        deviceToken: deviceToken.token,
        platform: deviceToken.platform
      });
      console.log('✅ Device token added event published');
    } catch (eventError) {
      console.error('❌ Failed to publish device token added event:', eventError);
    }

    return deviceToken;
  }

  async removeDeviceToken(userId: string, token: string) {
    await this.prisma.deviceToken.updateMany({
      where: {
        userId,
        token,
      },
      data: {
        isActive: false,
      },
    });

     try {
      await this.rabbitMQService.publish('user.device-token.removed', {
        userId,
        deviceToken: token
      });
      console.log('✅ Device token removed event published');
    } catch (eventError) {
      console.error('❌ Failed to publish device token removed event:', eventError);
    }
  }


  

  async getActiveDeviceTokens(userId: string) {
    return this.prisma.deviceToken.findMany({
      where: {
        userId,
        isActive: true,
      },
    });
  }
}