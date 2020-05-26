import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ManagePhraseFormComponent } from './manage-phrase-form.component';

describe('ManagePhraseFormComponent', () => {
  let component: ManagePhraseFormComponent;
  let fixture: ComponentFixture<ManagePhraseFormComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ManagePhraseFormComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ManagePhraseFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
