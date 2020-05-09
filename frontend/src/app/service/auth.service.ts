import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { share } from 'rxjs/operators';

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

  constructor(private http: HttpClient) { 
  }
  private playerUrl: string = environment.player_url;

  setPlayer(name:string, email:string, admin:boolean=false){
    this.player.name = name;
    this.player.email = email;
    this.player.admin = admin;
    this.isAuthed = true;
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

  identifyPlayer () {
    console.log("making the call")
    return this.http.get(this.playerUrl);
  }

}
