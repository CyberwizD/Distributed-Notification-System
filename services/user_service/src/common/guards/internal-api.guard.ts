import { Injectable, CanActivate, ExecutionContext, UnauthorizedException, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Observable } from 'rxjs';

@Injectable()
export class InternalApiGuard implements CanActivate {
  private readonly logger = new Logger(InternalApiGuard.name);

  constructor(private configService: ConfigService) {}

  canActivate(context: ExecutionContext): boolean | Promise<boolean> | Observable<boolean> {
    const request = context.switchToHttp().getRequest();
    const internalApiKey = request.headers['x-internal-api-key'];
    const expectedKey = this.configService.get('service.internalApiKey');

    this.logger.debug(`Received API Key: ${internalApiKey}`);
    this.logger.debug(`Expected API Key: ${expectedKey}`);
    this.logger.debug(`Headers: ${JSON.stringify(request.headers)}`);

    if (!internalApiKey) {
      this.logger.warn('No internal API key provided');
      throw new UnauthorizedException('No internal API key provided');
    }

    if (internalApiKey === expectedKey) {
      this.logger.debug('Internal API key validation successful');
      return true;
    }

    this.logger.warn(`Invalid internal API key. Received: ${internalApiKey}, Expected: ${expectedKey}`);
    throw new UnauthorizedException('Invalid internal API key');
  }
}