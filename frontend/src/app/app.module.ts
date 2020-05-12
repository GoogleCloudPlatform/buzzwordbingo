import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { BoardComponent } from './view/board/board.component';
import { ItemComponent } from './view/item/item.component';
import { HttpClientModule }    from '@angular/common/http';
import { environment } from "src/environments/environment";
import { AngularFireModule } from "@angular/fire";
import { AngularFirestoreModule } from "@angular/fire/firestore";
import { LoginComponent } from './view/login/login.component';
import { MessagesComponent } from './view/messages/messages.component';
import { ReplacePipe } from './replace.pipe';
import { AdminComponent } from './view/admin/admin.component';
import { ToolbarComponent } from './view/toolbar/toolbar.component';
import { MasterComponent } from './view/master/master.component';
import { AdminBoardsComponent } from './view/admin-boards/admin-boards.component';

@NgModule({
  declarations: [
    AppComponent,
    BoardComponent,
    ItemComponent,
    LoginComponent,
    ReplacePipe,
    MessagesComponent,
    AdminComponent,
    ToolbarComponent,
    MasterComponent,
    AdminBoardsComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    AngularFireModule.initializeApp(environment.firebaseConfig),
    AngularFirestoreModule,
    HttpClientModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
