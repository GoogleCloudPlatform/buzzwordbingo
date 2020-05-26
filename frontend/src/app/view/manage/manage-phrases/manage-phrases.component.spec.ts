import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ManagePhrasesComponent } from './manage-phrases.component';

describe('ManagePhrasesComponent', () => {
  let component: ManagePhrasesComponent;
  let fixture: ComponentFixture<ManagePhrasesComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ManagePhrasesComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ManagePhrasesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
