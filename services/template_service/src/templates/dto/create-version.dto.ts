import { IsString, IsBoolean, IsOptional } from 'class-validator';

export class CreateVersionDto {
  @IsString() subject: string;
  @IsString() body: string;
  @IsOptional() @IsBoolean() is_active?: boolean;
}
