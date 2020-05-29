import { Injectable } from '@angular/core';
import {Game} from "../service/game.service"
import  {Player} from "../service/auth.service"

@Injectable({
  providedIn: 'root'
})
export class LocalstorageService {

  constructor() { }

  setTheme(theme:string){
    localStorage.setItem("theme", theme);
  }

  getTheme():string{
    return localStorage.getItem('theme')
  }

  getGame():Game{
    return JSON.parse(localStorage.getItem('game'));
  }

  getPlayer(){
    return JSON.parse(localStorage.getItem('player'));
  }

  setPlayer(player:Player){
    localStorage.setItem('player', JSON.stringify(player));
  }

  clearGameData(){
    localStorage.removeItem("game");
    localStorage.removeItem("player");
  }
}
