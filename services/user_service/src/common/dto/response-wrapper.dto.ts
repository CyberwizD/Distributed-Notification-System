import { ApiProperty } from '@nestjs/swagger';

export class ResponseWrapper<T> {
  @ApiProperty()
  success: boolean;

  @ApiProperty()
  message: string;

  @ApiProperty()
  data?: T;

  @ApiProperty()
  timestamp: string;

  constructor(success: boolean, message: string, data?: T) {
    this.success = success;
    this.message = message;
    this.data = data;
    this.timestamp = new Date().toISOString();
  }

  static success<T>(message: string, data?: T): ResponseWrapper<T> {
    return new ResponseWrapper(true, message, data);
  }

  static error<T>(message: string): ResponseWrapper<T> {
    return new ResponseWrapper(false, message);
  }
}