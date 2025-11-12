import { Injectable } from '@nestjs/common';
import Handlebars from 'handlebars';

@Injectable()
export class SubstitutionService {
  render(template: string, variables: Record<string, any>): string {
    const compiled = Handlebars.compile(template);
    return compiled(variables);
  }
}
