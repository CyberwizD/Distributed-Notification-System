import {
  BadRequestException,
  Injectable,
  NotFoundException,
  // BadRequestException,
} from '@nestjs/common';
// import { PrismaService } from '../config/database.config';
import { SubstitutionService } from '../substitution/substitution.service';

import { CreateTemplateDto } from './dto/create-template.dto';
import { CreateVersionDto } from './dto/create-version.dto';
import { RenderTemplateDto } from './dto/render-template.dto';
import { PrismaService } from 'src/config/prisma.service';
import { UpdateTemplateDto } from './dto/update-template.dto';
import { ResponseWrapperDto } from 'src/common/dto/response-wrapper.dto';
import { Template, TemplateVersion } from '@prisma/client';

@Injectable()
export class TemplatesService {
  constructor(
    private prisma: PrismaService,
    private substitutionService: SubstitutionService,
  ) {}

  private handlePrismaError(error: any): string {
    if (error.code === 'P2002') {
      return `Duplicate field value: ${error.meta.target.join(', ')}`;
    }
    return error.message;
  }

  private async _getTemplateById(id: string): Promise<Template> {
    const template = await this.prisma.template.findUnique({ where: { id } });
    if (!template) {
      throw new NotFoundException('Template not found');
    }
    return template;
  }

  private async _getTemplateByIdentifier(
    identifier: string,
    locale?: string,
  ): Promise<Template> {
    const where: any = {
      OR: [{ id: identifier }, { name: identifier }],
    };
    if (locale) {
      where.language = locale;
    }
    const template = await this.prisma.template.findFirst({ where });
    if (!template) {
      throw new NotFoundException('Template not found');
    }
    return template;
  }

  private async _getActiveVersion(
    templateId: string,
  ): Promise<TemplateVersion> {
    const version = await this.prisma.templateVersion.findFirst({
      where: { templateId, isActive: true },
    });
    if (!version) {
      throw new NotFoundException('Active template version not found');
    }
    return version;
  }

  async getTemplates(): Promise<ResponseWrapperDto<Template[]>> {
    try {
      const templates = await this.prisma.template.findMany({
        include: {
          versions: true,
        },
      });
      return new ResponseWrapperDto({
        success: true,
        message: 'Templates retrieved successfully',
        data: templates,
      });
    } catch (error) {
      return new ResponseWrapperDto({
        success: false,
        message: 'Failed to retrieve templates',
        error:
          error instanceof Error ? error.message : 'An unknown error occurred',
      });
    }
  }

  async getTemplateById(
    id: string,
  ): Promise<ResponseWrapperDto<Template | null>> {
    try {
      const template = await this._getTemplateById(id);
      return new ResponseWrapperDto({
        success: true,
        message: 'Template retrieved successfully',
        data: template,
      });
    } catch (error) {
      return new ResponseWrapperDto({
        success: false,
        message: 'Failed to retrieve template',
        error:
          error instanceof Error ? error.message : 'An unknown error occurred',
      });
    }
  }

  async updateTemplate(
    id: string,
    dto: UpdateTemplateDto,
  ): Promise<ResponseWrapperDto<Template | null>> {
    try {
      await this._getTemplateById(id);
      const { versions, ...templateData } = dto;
      const updatedTemplate = await this.prisma.template.update({
        where: { id },
        data: templateData,
      });
      return new ResponseWrapperDto({
        success: true,
        message: 'Template updated successfully',
        data: updatedTemplate,
      });
    } catch (error) {
      return new ResponseWrapperDto({
        success: false,
        message: 'Failed to update template',
        error: this.handlePrismaError(error),
      });
    }
  }

  async deleteTemplate(
    id: string,
  ): Promise<ResponseWrapperDto<Template | null>> {
    try {
      await this._getTemplateById(id);
      const deletedTemplate = await this.prisma.template.delete({
        where: { id },
      });
      return new ResponseWrapperDto({
        success: true,
        message: 'Template deleted successfully',
        data: deletedTemplate,
      });
    } catch (error) {
      return new ResponseWrapperDto({
        success: false,
        message: 'Failed to delete template',
        error:
          error instanceof Error ? error.message : 'An unknown error occurred',
      });
    }
  }

  async createTemplate(
    dto: CreateTemplateDto,
  ): Promise<ResponseWrapperDto<Template>> {
    try {
      const { name, language, type, versions } = dto;
      const versionsToCreate = Array.isArray(versions)
        ? versions
        : versions
          ? [versions]
          : [];
      const template = await this.prisma.template.create({
        data: {
          name,
          language,
          type,
          versions: {
            create: versionsToCreate.map((version, index) => ({
              ...version,
              isActive: index === 0,
            })),
          },
        },
        include: {
          versions: true,
        },
      });
      return new ResponseWrapperDto({
        success: true,
        message: 'Template created successfully',
        data: template,
      });
    } catch (error) {
      return new ResponseWrapperDto({
        success: false,
        message: 'Failed to create template',
        error: this.handlePrismaError(error),
      });
    }
  }

  async createVersion(
    templateId: string,
    dto: CreateVersionDto,
  ): Promise<ResponseWrapperDto<TemplateVersion>> {
    try {
      await this._getTemplateById(templateId);

      if (dto.is_active) {
        await this.prisma.templateVersion.updateMany({
          where: { templateId },
          data: { isActive: false },
        });
      }

      const newVersion = await this.prisma.templateVersion.create({
        data: {
          ...dto,
          templateId,
        },
      });

      return new ResponseWrapperDto({
        success: true,
        message: 'Template version created successfully',
        data: newVersion,
      });
    } catch (error) {
      return new ResponseWrapperDto({
        success: false,
        message: 'Failed to create template version',
        error:
          error instanceof Error ? error.message : 'An unknown error occurred',
      });
    }
  }

  async setActiveVersion(
    templateId: string,
    versionId: string,
  ): Promise<ResponseWrapperDto<TemplateVersion>> {
    try {
      await this._getTemplateById(templateId);
      const version = await this.prisma.templateVersion.findUnique({
        where: { id: versionId },
      });
      if (!version) {
        throw new NotFoundException('Template version not found');
      }

      await this.prisma.templateVersion.updateMany({
        where: { templateId },
        data: { isActive: false },
      });

      const updatedVersion = await this.prisma.templateVersion.update({
        where: { id: versionId },
        data: { isActive: true },
      });

      return new ResponseWrapperDto({
        success: true,
        message: 'Template version activated successfully',
        data: updatedVersion,
      });
    } catch (error) {
      return new ResponseWrapperDto({
        success: false,
        message: 'Failed to activate template version',
        error:
          error instanceof Error ? error.message : 'An unknown error occurred',
      });
    }
  }

  async getActiveVersion(
    templateIdentifier: string,
    locale?: string,
  ): Promise<ResponseWrapperDto<TemplateVersion | null>> {
    try {
      const template = await this._getTemplateByIdentifier(
        templateIdentifier,
        locale,
      );
      const version = await this._getActiveVersion(template.id);
      return new ResponseWrapperDto({
        success: true,
        message: 'Active template version retrieved successfully',
        data: version,
      });
    } catch (error) {
      return new ResponseWrapperDto({
        success: false,
        message: 'Failed to retrieve active template version',
        error:
          error instanceof Error ? error.message : 'An unknown error occurred',
      });
    }
  }

  async renderTemplate(
    dto: RenderTemplateDto,
  ): Promise<ResponseWrapperDto<{ subject: string; body: string } | null>> {
    try {
      const identifier = dto.template_id ?? dto.template_code;
      if (!identifier) {
        throw new BadRequestException(
          'template_id or template_code is required',
        );
      }
      const { variables } = dto;
      const template = await this._getTemplateByIdentifier(identifier);
      const version = await this._getActiveVersion(template.id);

      const rendered = {
        subject: this.substitutionService.render(version.subject, variables),
        body: this.substitutionService.render(version.body, variables),
      };

      return new ResponseWrapperDto({
        success: true,
        message: 'Template rendered successfully',
        data: rendered,
      });
    } catch (error) {
      return new ResponseWrapperDto({
        success: false,
        message: 'Failed to render template',
        error:
          error instanceof Error ? error.message : 'An unknown error occurred',
      });
    }
  }
}
