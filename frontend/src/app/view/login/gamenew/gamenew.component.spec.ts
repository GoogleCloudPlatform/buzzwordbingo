import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { GamenewComponent } from './gamenew.component';

describe('GamenewComponent', () => {
  let component: GamenewComponent;
  let fixture: ComponentFixture<GamenewComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ GamenewComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(GamenewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
