import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { LoginComponent } from './view/login/login.component';
import { BoardComponent } from './view/board/board.component';
import { AdminComponent } from './view/admin/admin.component';
import { AuthGuard } from '../app/auth/auth.guard';


const routes: Routes = [
  { path: '', redirectTo: '/login', pathMatch: 'full' },
  { path: 'login/:id', component: LoginComponent },
  { path: 'game/:id', component: BoardComponent },
  { path: 'admin', component: AdminComponent, canActivate: [AuthGuard] },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
