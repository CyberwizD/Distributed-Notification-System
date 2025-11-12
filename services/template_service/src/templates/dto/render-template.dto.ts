import { IsString, IsObject, IsOptional } from 'class-validator';

export class RenderTemplateDto {
  @IsString()
  @IsOptional()
  template_id?: string;

  @IsString()
  @IsOptional()
  template_code?: string;

  @IsObject()
  variables: Record<string, any>;
}
