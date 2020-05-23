import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AdminPhrasesComponent } from './admin-phrases.component';

describe('AdminPhrasesComponent', () => {
  let component: AdminPhrasesComponent;
  let fixture: ComponentFixture<AdminPhrasesComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ AdminPhrasesComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AdminPhrasesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
