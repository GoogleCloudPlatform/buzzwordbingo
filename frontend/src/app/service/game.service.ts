import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';
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

@Injectable({
  providedIn: 'root'
})
export class GameService {

  constructor(private http: HttpClient) { }
  private boardUrl: string = environment.board_url;
  private recordUrl: string = environment.record_url;

  getBoard (email:string): Observable<Board> {
    return this.http.get<Board>(this.boardUrl +"?email="+email);
  }

  record (pid:string, bid:string) {
    console.log("Record", pid, bid)
    return this.http.get(this.recordUrl + "?p="+pid + "&b=" + bid).subscribe(val=>console.log(val));
  }

}
