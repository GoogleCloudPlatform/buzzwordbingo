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
  text:string
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
  id:string
  
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
  private hostUrl: string = environment.host_url;

  getBoard (name:string, g:string): Observable<Board> {
    if (name == "undefined") return
    return this.http.get<Board>(this.hostUrl +"/api/board?name="+name + "&g="+g).pipe(share());
  }

  isAdmin (): Observable<boolean> {
    return this.http.get<boolean>(this.hostUrl+ "/api/player/isadmin");
  }


  record (pid:string, gid:string,  bid:string) {
    return this.http.get(this.hostUrl + "/api/record?p="+pid + "&g=" + gid + "&b=" + bid).subscribe();
  }

  resetboard (bid:string, gid:string) {
    return this.http.get(this.hostUrl + "/api/board/delete?b=" + bid + "&g=" + gid).subscribe();
  }

  newGame (name:string) {
    return this.http.get(this.hostUrl + "/api/game/new?name=" + name);
  }

  getGame (gid:string) {
    return this.http.get(this.hostUrl +  "/api/game?g="+gid).pipe(share());
  }

  getGamesForPlayer(){
    return this.http.get(this.hostUrl +  "/api/game/list").pipe(share());
  }

}
