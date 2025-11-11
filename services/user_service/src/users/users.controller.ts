import {
  Controller,
  Get,
  Post,
  Body,
  Patch,
  Param,
  Delete,
  Query,
  UseGuards,
  ParseIntPipe,
  DefaultValuePipe,
} from '@nestjs/common';
import { ApiTags, ApiOperation, ApiResponse, ApiBearerAuth, ApiQuery } from '@nestjs/swagger';
import { JwtAuthGuard } from '../common/guards/jwt-auth.guard';
import { UsersService } from './users.service';
import { CreateUserDto } from './dto/create-user.dto';
import { UpdateUserDto } from './dto/update-user.dto';
import { UpdatePreferenceDto } from './dto/update-preference.dto';
import { AddDeviceTokenDto, RemoveDeviceTokenDto } from './dto/device-token.dto';
import { ResponseWrapper } from '../common/dto/response-wrapper.dto';

@ApiTags('Users')
@ApiBearerAuth()
@UseGuards(JwtAuthGuard)
@Controller('users')
export class UsersController {
  constructor(private readonly usersService: UsersService) {}

  @Post()
  @ApiOperation({ summary: 'Create a new user' })
  @ApiResponse({ status: 201, description: 'User created successfully' })
  async create(@Body() createUserDto: CreateUserDto) {
    const user = await this.usersService.create(createUserDto);
    return ResponseWrapper.success('User created successfully', user);
  }

  @Get()
  @ApiOperation({ summary: 'Get all users with pagination' })
  @ApiQuery({ name: 'page', required: false, type: Number })
  @ApiQuery({ name: 'limit', required: false, type: Number })
  async findAll(
    @Query('page', new DefaultValuePipe(1), ParseIntPipe) page: number,
    @Query('limit', new DefaultValuePipe(10), ParseIntPipe) limit: number,
  ) {
    const result = await this.usersService.findAll(page, limit);
    return ResponseWrapper.success('Users retrieved successfully', result);
  }

  @Get(':id')
  @ApiOperation({ summary: 'Get user by ID' })
  @ApiResponse({ status: 200, description: 'User found' })
  @ApiResponse({ status: 404, description: 'User not found' })
  async findOne(@Param('id') id: string) {
    const user = await this.usersService.findOne(id);
    return ResponseWrapper.success('User retrieved successfully', user);
  }

  @Patch(':id')
  @ApiOperation({ summary: 'Update user' })
  async update(@Param('id') id: string, @Body() updateUserDto: UpdateUserDto) {
    const user = await this.usersService.update(id, updateUserDto);
    return ResponseWrapper.success('User updated successfully', user);
  }

  @Delete(':id')
  @ApiOperation({ summary: 'Delete user (soft delete)' })
  async remove(@Param('id') id: string) {
    await this.usersService.remove(id);
    return ResponseWrapper.success('User deleted successfully');
  }

  @Get(':id/preferences')
  @ApiOperation({ summary: 'Get user preferences' })
  async getPreferences(@Param('id') id: string) {
    const preferences = await this.usersService.getPreferences(id);
    return ResponseWrapper.success('Preferences retrieved successfully', preferences);
  }

  @Patch(':id/preferences')
  @ApiOperation({ summary: 'Update user preferences' })
  async updatePreferences(
    @Param('id') id: string,
    @Body() updatePreferenceDto: UpdatePreferenceDto,
  ) {
    const preferences = await this.usersService.updatePreferences(id, updatePreferenceDto);
    return ResponseWrapper.success('Preferences updated successfully', preferences);
  }

  // Internal endpoints for other services
  @Get(':id/contact')
  @ApiOperation({ summary: 'Get user contact info (for internal use)' })
  async getContactInfo(@Param('id') id: string) {
    const contactInfo = await this.usersService.getUserContactInfo(id);
    return ResponseWrapper.success('Contact info retrieved successfully', contactInfo);
  }

  @Get(':id/can-receive-email')
  @ApiOperation({ summary: 'Check if user can receive emails' })
  async canReceiveEmail(@Param('id') id: string) {
    const canReceive = await this.usersService.canReceiveEmail(id);
    return ResponseWrapper.success('Check completed', { canReceiveEmail: canReceive });
  }

  @Get(':id/can-receive-push')
  @ApiOperation({ summary: 'Check if user can receive push notifications' })
  async canReceivePush(@Param('id') id: string) {
    const canReceive = await this.usersService.canReceivePush(id);
    return ResponseWrapper.success('Check completed', { canReceivePush: canReceive });
  }

  // Device token management
  @Post(':id/device-tokens')
  @ApiOperation({ summary: 'Add device token for push notifications' })
  @ApiResponse({ status: 201, description: 'Device token added successfully' })
  async addDeviceToken(
    @Param('id') id: string,
    @Body() addDeviceTokenDto: AddDeviceTokenDto,
  ) {
    const deviceToken = await this.usersService.addDeviceToken(
      id,
      addDeviceTokenDto.token,
      addDeviceTokenDto.platform,
    );
    return ResponseWrapper.success('Device token added successfully', deviceToken);
  }

  @Delete(':id/device-tokens')
  @ApiOperation({ summary: 'Remove device token' })
  @ApiResponse({ status: 200, description: 'Device token removed successfully' })
  async removeDeviceToken(
    @Param('id') id: string,
    @Body() removeDeviceTokenDto: RemoveDeviceTokenDto,
  ) {
    await this.usersService.removeDeviceToken(id, removeDeviceTokenDto.token);
    return ResponseWrapper.success('Device token removed successfully');
  }

  @Get(':id/device-tokens')
  @ApiOperation({ summary: 'Get all active device tokens for user' })
  @ApiResponse({ status: 200, description: 'Device tokens retrieved successfully' })
  async getDeviceTokens(@Param('id') id: string) {
    const tokens = await this.usersService.getActiveDeviceTokens(id);
    return ResponseWrapper.success('Device tokens retrieved successfully', tokens);
  }
}