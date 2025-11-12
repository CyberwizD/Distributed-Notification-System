import {
  Injectable,
  NestInterceptor,
  ExecutionContext,
  CallHandler,
} from '@nestjs/common';
import { Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
import { v4 as uuidv4 } from 'uuid';

@Injectable()
export class CorrelationIdInterceptor implements NestInterceptor {
  intercept(context: ExecutionContext, next: CallHandler): Observable<any> {
    const request = context.switchToHttp().getRequest();
    const response = context.switchToHttp().getResponse();
    
    let correlationId = request.headers['x-correlation-id'];
    if (!correlationId) {
      correlationId = uuidv4();
    }
    
    request.correlationId = correlationId;
    response.setHeader('x-correlation-id', correlationId);
    
    return next.handle();
  }
}