import { Injectable } from '@angular/core';

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
  private isAuthed:boolean=false;

  constructor() { }


  setPlayer(name:string, email:string, admin:boolean=false){
    this.player.name = name;
    this.player.email = email;
    this.player.admin = admin;
    this.isAuthed = true;
  } 

  getPlayer():Player{
    return this.player;
  }

  isAuth():boolean{
    return this.isAuthed;
  }

}
