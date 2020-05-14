import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AdminPhraseFormComponent } from './admin-phrase-form.component';

describe('AdminPhraseFormComponent', () => {
  let component: AdminPhraseFormComponent;
  let fixture: ComponentFixture<AdminPhraseFormComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ AdminPhraseFormComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AdminPhraseFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
