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
	audience:string   
}

@Injectable({
  providedIn: 'root'
})
export class GameService {

  game:any = new Game;

  constructor(private http: HttpClient) { 
  }
  private boardUrl: string = environment.board_url;
  private recordUrl: string = environment.record_url;
  private gameActiveUrl: string = environment.game_active_url;

  getBoard (email:string, name:string): Observable<Board> {
    if (email == "undefined") return
    return this.http.get<Board>(this.boardUrl +"?email="+email+"&name="+name).pipe(share());
  }

  getActiveGame () {
    console.log("active game called")
    return this.http.get(this.gameActiveUrl).subscribe(val=>{this.game=val; console.log(val)});
  }

  record (pid:string, bid:string) {
    return this.http.get(this.recordUrl + "?p="+pid + "&b=" + bid).subscribe();
  }

}
