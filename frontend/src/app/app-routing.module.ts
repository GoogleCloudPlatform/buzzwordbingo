/**
 * Copyright 2020 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { LoginComponent } from './view/login/login.component';
import { BoardComponent } from './view/board/board.component';
import { AdminComponent } from './view/admin/admin.component';
import { AuthGuard } from '../app/auth/auth.guard';
import { GamepickerComponent } from './view/login/gamepicker/gamepicker.component';
import { ManageComponent } from './view/manage/manage.component';


const routes: Routes = [
  { path: '', redirectTo: '/login', pathMatch: 'full' },
  { path: 'login/:id', component: LoginComponent },
  { path: 'invite/:id', component: LoginComponent },
  { path: 'login', component: LoginComponent },
  { path: 'game/:id', component: BoardComponent },
  { path: 'gamepicker', component: GamepickerComponent },
  { path: 'gamenew', component:  GamepickerComponent},
  { path: 'manage/:id', component:  ManageComponent},
  { path: 'admin', component: AdminComponent, canActivate: [AuthGuard] },
  { path: 'admin/:id', component: AdminComponent, canActivate: [AuthGuard] },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
