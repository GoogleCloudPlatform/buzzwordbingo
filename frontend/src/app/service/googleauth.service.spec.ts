import { TestBed } from '@angular/core/testing';

import { GoogleauthService } from './googleauth.service';

describe('GoogleauthService', () => {
  let service: GoogleauthService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(GoogleauthService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
