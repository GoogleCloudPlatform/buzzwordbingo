import { Injectable } from '@angular/core';
import { share } from 'rxjs/operators';
import { Observable, of  } from 'rxjs';
import { Player} from '../service/auth.service'
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { environment } from '../../environments/environment';

export class Board{
  id:string
  game:string
  player:Player
  phrases:Phrase[]
  bingodeclared:boolean=false
}

export class Record{
  phrase:Phrase
  players:Player[]
}

export class Phrase{
  id:number
  phrase:string
  selected:boolean
  tid:string
}

export class Game  {
	id:string 
	name:string 
	active:boolean   
}

export class Message  {
  text:string 
  bingo:boolean
  audience:string[]  
  operation:string 
  
  public isAudience(email:string):boolean{
    this.audience.forEach(function(aud:string) {			
      if (aud == email){
        return true;
      }			
    })	
    return false;	
  }

  public isAll():boolean{
    this.audience.forEach(function(aud:string) {			
      if (aud == "all"){
        return true;
      }			
    })
    return false;	
  }
}

@Injectable({
  providedIn: 'root'
})
export class GameService {

  game:any = new Game;

  constructor(private http: HttpClient) { 
    let game = JSON.parse(localStorage.getItem('game'));
    if (game != null){
      this.game = game;
    }
  }
  private boardUrl: string = environment.board_url;
  private recordUrl: string = environment.record_url;
  private gameActiveUrl: string = environment.game_active_url;
  private adminUrl: string = environment.admin_url;
  private resetUrl: string = environment.reset_url;

  getBoard (email:string, name:string): Observable<Board> {
    if (email == "undefined") return
    return this.http.get<Board>(this.boardUrl +"?email="+email+"&name="+name).pipe(share());
  }

  isAdmin (email:string): Observable<boolean> {
    if (email == "undefined") return
    return this.http.get<boolean>(this.adminUrl +"?email="+email);
  }

  getActiveGame () {
    return this.http.get(this.gameActiveUrl).pipe(share()).subscribe(val=>{this.game=val; localStorage.setItem('game', JSON.stringify(val));});
  }

  record (pid:string, bid:string) {
    return this.http.get(this.recordUrl + "?p="+pid + "&b=" + bid).subscribe();
  }

  resetboard (bid:string) {
    return this.http.get(this.resetUrl + "?b=" + bid).subscribe(val=>console.log("reset result",val));
  }

}
