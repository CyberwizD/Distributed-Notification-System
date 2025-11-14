import { Controller, Get, Param, UseGuards } from '@nestjs/common';
import { ApiOperation, ApiTags } from '@nestjs/swagger';
import { InternalApiGuard } from '../common/guards/internal-api.guard';
import { UsersService } from './users.service';
import { ResponseWrapper } from '../common/dto/response-wrapper.dto';

@ApiTags('Internal Users')
@UseGuards(InternalApiGuard)
@Controller('internal/users')
export class InternalUsersController {
  constructor(private readonly usersService: UsersService) {}

  @Get(':id/notification-profile')
  @ApiOperation({ summary: 'Get notification profile for downstream services' })
  async getNotificationProfile(@Param('id') id: string) {
    const contactInfo = await this.usersService.getUserContactInfo(id);
    return ResponseWrapper.success('Notification profile retrieved', {
      userId: id,
      email: contactInfo.email,
      preferences: contactInfo.preferences,
      deviceTokens: contactInfo.deviceTokens,
    });
  }
}
