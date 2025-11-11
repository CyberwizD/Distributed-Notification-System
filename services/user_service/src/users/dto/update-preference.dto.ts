import { ApiProperty } from '@nestjs/swagger';
import { IsBoolean, IsString, IsOptional, IsEnum } from 'class-validator';

export class UpdatePreferenceDto {
  @ApiProperty({ required: false })
  @IsBoolean()
  @IsOptional()
  emailEnabled?: boolean;

  @ApiProperty({ required: false })
  @IsBoolean()
  @IsOptional()
  pushEnabled?: boolean;

  @ApiProperty({ required: false })
  @IsBoolean()
  @IsOptional()
  smsEnabled?: boolean;

  @ApiProperty({ enum: ['immediate', 'daily', 'weekly', 'never'], required: false })
  @IsEnum(['immediate', 'daily', 'weekly', 'never'])
  @IsOptional()
  emailFrequency?: string;

  @ApiProperty({ required: false })
  @IsString()
  @IsOptional()
  language?: string;

  @ApiProperty({ required: false })
  @IsString()
  @IsOptional()
  timezone?: string;

  @ApiProperty({ required: false })
  @IsBoolean()
  @IsOptional()
  marketingEmails?: boolean;

  @ApiProperty({ required: false })
  @IsBoolean()
  @IsOptional()
  securityEmails?: boolean;
}