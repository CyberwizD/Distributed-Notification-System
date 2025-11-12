import { Injectable, UnauthorizedException, ConflictException, BadRequestException } from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { ConfigService } from '@nestjs/config';
import * as bcrypt from 'bcryptjs';
import { PrismaService } from '../prisma/prisma.service';
import { RabbitMQService } from '../common/rabbitmq/rabbitmq.service';
import { RegisterDto } from './dto/register.dto';
import { LoginDto } from './dto/login.dto';

@Injectable()
export class AuthService {
  constructor(
    private prisma: PrismaService,
    private jwtService: JwtService,
    private configService: ConfigService,
    private rabbitMQService: RabbitMQService,
  ) { }

  async register(registerDto: RegisterDto) {
    const { email, password, name } = registerDto;

    // Check if user exists
    const existingUser = await this.prisma.user.findUnique({
      where: { email },
    });

    if (existingUser) {
      throw new ConflictException('User with this email already exists');
    }

    // Hash password
    const hashedPassword = await bcrypt.hash(password, 12);

    try {
      // Create user with preferences
      const user = await this.prisma.user.create({
        data: {
          email,
          password: hashedPassword,
          name,
          preferences: {
            create: {
              emailEnabled: true,
              pushEnabled: true,
              smsEnabled: false,
              emailFrequency: 'immediate',
              language: 'en',
              timezone: 'UTC',
              marketingEmails: false,
              securityEmails: true,
            },
          },
        },
        include: {
          preferences: true,
        },
      });


      // Remove password from response
      const { password: _, ...userWithoutPassword } = user;

      // Publish user registered event
      try {
        await this.rabbitMQService.publish('user.registered', {
          userId: user.id,
          email: user.email,
          name: user.name,
          preferences: user.preferences
        });
        console.log('✅ User registered event published');
      } catch (eventError) {
        console.error('❌ Failed to publish user registered event:', eventError);
        // Don't throw error - user creation should succeed even if event fails
      }

      // Generate token
      const token = this.generateToken(user.id);

      return {
        user: userWithoutPassword,
        token,
      };
    } catch (error) {
      throw new BadRequestException('Failed to create user');
    }
  }

  async login(loginDto: LoginDto) {
    const { email, password } = loginDto;

    // Find user
    const user = await this.prisma.user.findUnique({
      where: { email },
      include: { preferences: true },
    });

    if (!user || !user.isActive) {
      throw new UnauthorizedException('Invalid credentials');
    }

    // Check password
    const isPasswordValid = await bcrypt.compare(password, user.password);
    if (!isPasswordValid) {
      throw new UnauthorizedException('Invalid credentials');
    }

    // Generate token
    const token = this.generateToken(user.id);

    return {
      user: {
        id: user.id,
        email: user.email,
        name: user.name,
        preferences: user.preferences,
      },
      token,
    };
  }

  async validateUser(userId: string) {
    return this.prisma.user.findUnique({
      where: { id: userId },
      include: { preferences: true },
    });
  }

  private generateToken(userId: string): string {
    const payload = { sub: userId };
    return this.jwtService.sign(payload, {
      secret: this.configService.get('jwt.secret'),
      expiresIn: this.configService.get('jwt.expiresIn'),
    });
  }

   private isValidEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }
}