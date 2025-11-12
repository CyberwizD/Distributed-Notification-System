import { PaginationMetaDto } from './pagination-meta.dto';

export class ResponseWrapperDto<T> {
  success: boolean;
  message: string;
  data?: T;
  error?: string | null;
  meta?: PaginationMetaDto | null;

  constructor(partial: Partial<ResponseWrapperDto<T>>) {
    Object.assign(this, partial);
  }
}
