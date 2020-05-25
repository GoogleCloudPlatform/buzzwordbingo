import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ManageMasterComponent } from './manage-master.component';

describe('ManageMasterComponent', () => {
  let component: ManageMasterComponent;
  let fixture: ComponentFixture<ManageMasterComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ManageMasterComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ManageMasterComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
