import {
  IsArray,
  IsEnum,
  IsObject,
  IsString,
  ValidateNested,
} from 'class-validator';
import { Type } from 'class-transformer';

class TemplateVersionDto {
  @IsString()
  subject: string;

  @IsString()
  body: string;
}

export class CreateTemplateDto {
  @IsString() name: string;

  @IsString()
  @IsEnum(['push', 'email'])
  type: string;

  @IsString() language: string;

  @IsArray()
  @ValidateNested({ each: true })
  @Type(() => TemplateVersionDto)
  versions: TemplateVersionDto[];
}
