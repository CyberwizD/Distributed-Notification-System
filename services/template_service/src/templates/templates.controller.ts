import {
  Body,
  Controller,
  Delete,
  Get,
  Param,
  Patch,
  Post,
  Query,
} from '@nestjs/common';
import { TemplatesService } from './templates.service';
import { CreateTemplateDto } from './dto/create-template.dto';
import { UpdateTemplateDto } from './dto/update-template.dto';
import { CreateVersionDto } from './dto/create-version.dto';
import { RenderTemplateDto } from './dto/render-template.dto';

@Controller('api/v1/templates')
export class TemplatesController {
  constructor(private readonly templatesService: TemplatesService) {}

  @Get()
  getTemplates() {
    return this.templatesService.getTemplates();
  }

  @Post()
  createTemplate(@Body() dto: CreateTemplateDto) {
    return this.templatesService.createTemplate(dto);
  }

  @Get(':id')
  getTemplateById(@Param('id') id: string) {
    return this.templatesService.getTemplateById(id);
  }

  @Patch(':id')
  updateTemplate(@Param('id') id: string, @Body() dto: UpdateTemplateDto) {
    return this.templatesService.updateTemplate(id, dto);
  }

  @Delete(':id')
  deleteTemplate(@Param('id') id: string) {
    return this.templatesService.deleteTemplate(id);
  }

  @Post(':id/versions')
  createVersion(@Param('id') id: string, @Body() dto: CreateVersionDto) {
    return this.templatesService.createVersion(id, dto);
  }

  @Patch(':id/versions/:versionId/activate')
  setActiveVersion(
    @Param('id') id: string,
    @Param('versionId') versionId: string,
  ) {
    return this.templatesService.setActiveVersion(id, versionId);
  }

  @Get(':id/active')
  getActiveVersion(@Param('id') id: string, @Query('locale') locale?: string) {
    return this.templatesService.getActiveVersion(id, locale);
  }

  @Post('render')
  renderTemplate(@Body() dto: RenderTemplateDto) {
    return this.templatesService.renderTemplate(dto);
  }
}
