export class PaginationMetaDto {
  total: number;
  limit: number;
  page: number;
  total_pages: number;
  has_next: boolean;
  has_previous: boolean;

  constructor(total: number, limit: number, page: number) {
    this.total = total;
    this.limit = limit;
    this.page = page;
    this.total_pages = Math.ceil(total / limit);
    this.has_next = page < this.total_pages;
    this.has_previous = page > 1;
  }
}
