import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AdminBoardsComponent } from './admin-boards.component';

describe('AdminBoardsComponent', () => {
  let component: AdminBoardsComponent;
  let fixture: ComponentFixture<AdminBoardsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ AdminBoardsComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AdminBoardsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
