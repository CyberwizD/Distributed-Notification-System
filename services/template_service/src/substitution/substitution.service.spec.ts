import { Test, TestingModule } from '@nestjs/testing';
import { SubstitutionService } from './substitution.service';

describe('SubstitutionService', () => {
  let service: SubstitutionService;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [SubstitutionService],
    }).compile();

    service = module.get<SubstitutionService>(SubstitutionService);
  });

  it('should be defined', () => {
    expect(service).toBeDefined();
  });
});
