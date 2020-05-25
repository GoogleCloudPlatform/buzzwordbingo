import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ManageBoardsComponent } from './manage-boards.component';

describe('ManageBoardsComponent', () => {
  let component: ManageBoardsComponent;
  let fixture: ComponentFixture<ManageBoardsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ManageBoardsComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ManageBoardsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
