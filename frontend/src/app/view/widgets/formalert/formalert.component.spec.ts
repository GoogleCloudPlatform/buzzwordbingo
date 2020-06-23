import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { FormalertComponent } from './formalert.component';

describe('FormalertComponent', () => {
  let component: FormalertComponent;
  let fixture: ComponentFixture<FormalertComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ FormalertComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(FormalertComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
