import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { share } from 'rxjs/operators';
import {GameService, Game} from '../service/game.service'

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
  private isAdministrator:boolean=false;
  private playerUrl: string = environment.player_url;

  constructor(private http: HttpClient, public game:GameService) { 
    let player = JSON.parse(localStorage.getItem('player'));
    if (player != null){
      this.setPlayer(player.name, player.email)
    }
  }
  

  setPlayer(name:string, email:string, admin:boolean=false){
    this.player.name = name;
    this.player.email = email;
    this.player.admin = admin;
    this.isAuthed = true;
    localStorage.setItem('player', JSON.stringify(this.player));
    this.game.isAdmin(email).pipe(share()).subscribe(val=>{this.isAdministrator = val; console.log("admin",val)})
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

  identifyPlayer () {
    return this.http.get(this.playerUrl);
  }

}
