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

import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { share } from 'rxjs/operators';
import {GameService, Game} from '../service/game.service'
import {Router} from '@angular/router';
import { LocalstorageService } from './localstorage.service';
import { GoogleAuthService } from './googleauth.service';

export class Player{
  name:string
  email:string
  admin:boolean
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {

  private player:Player= new Player;
  private identity:any= new Player;
  private isAuthed:boolean=false;
  private isGameAdministrator:boolean=false;
  private isAdministrator:boolean=false;
  private hostUrl: string = environment.host_url;

  constructor(private http: HttpClient, 
              public game:GameService, 
              private localStorageService:LocalstorageService, 
              private googleAuth:GoogleAuthService,
              private router: Router) { 
    let player = localStorageService.getPlayer();
    if (player != null){
      this.setPlayer(player.name, player.email)
    }
  }
  

  setPlayer(name:string, email:string, admin:boolean=false){
    this.player.name = name;
    this.player.email = email;
    this.player.admin = admin;
    this.isAuthed = true;
    this.localStorageService.setPlayer(this.player);
    this.game.isAdmin().pipe(share()).subscribe(val=>{this.isAdministrator = val})
    this.googleAuth.login();

  } 

  getPlayer():Player{
    return this.player;
  }

  getIdentifiedEmail():string{
    return this.identity.email;
  }

  isAuth():boolean{
    return this.isAuthed;
  }

  isAdmin():boolean{
    return this.isAdministrator;
  }

  checkGameAdmin(gid:string){
    this.game.isGameAdmin(gid).pipe(share()).subscribe(val=>{this.isGameAdministrator = val})
  }

  isGameAdmin():boolean{
    return this.isGameAdministrator;
  }

  identifyPlayer () {
    return this.http.get(this.hostUrl + "/api/player/identify");
  }

  logout (reason:string="logged out") {
    console.log("logged out, reason:", reason )
    this.localStorageService.clearGameData();
    this.googleAuth.logout()
    this.router.navigateByUrl('/login');
    return 
  }

}
