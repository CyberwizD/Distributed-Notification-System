import { Module } from '@nestjs/common';
import { UsersService } from './users.service';
import { UsersController } from './users.controller';
import { InternalUsersController } from './internal-users.controller';

@Module({
  controllers: [UsersController, InternalUsersController],
  providers: [UsersService],
})
export class UsersModule {}
