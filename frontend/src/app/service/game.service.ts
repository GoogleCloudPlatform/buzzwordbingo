import { Injectable } from '@angular/core';
import { share } from 'rxjs/operators';
import { Observable, of  } from 'rxjs';
import { Player} from '../service/auth.service'
import { HttpClient, HttpHeaders} from '@angular/common/http';
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
  master:Master   
  admins:Player[]
  players:Player[]
}

export class Master {
  records:Record[]
}

export class Message  {
  text:string 
  bingo:boolean
  audience:string[]  
  operation:string 
  id:string
  received:boolean
  
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
    return this.http.get<boolean>(this.hostUrl+ "/api/player/isadmin").pipe(share());
  }

  isGameAdmin (gid:string): Observable<boolean> {
    return this.http.get<boolean>(this.hostUrl+ "/api/game/isadmin?g=" + gid).pipe(share());
  }

  record (pid:string, gid:string,  bid:string) {
    return this.http.get(this.hostUrl + "/api/record?p="+pid + "&g=" + gid + "&b=" + bid);
  }

  resetboard (bid:string, gid:string) {
    let url = `${this.hostUrl}/api/board/delete?b=${bid}&g=${gid}`
    console.log(url);
    return this.http.delete(url).subscribe();
  }

  newGame (name:string) {
    return this.http.get(this.hostUrl + "/api/game/new?name=" + name);
  }

  getGame (gid:string) {
    return this.http.get(this.hostUrl +  "/api/game?g="+gid).pipe(share());
  }

  deactivateGame (gid:string) {
    return this.http.get(this.hostUrl +  "/api/game/deactivate?g="+gid).pipe(share());
  }

  getGamesForPlayer(){
    return this.http.get(this.hostUrl +  "/api/game/list").pipe(share());
  }

  updateMasterPhrase(phrase:Phrase){
    let url = `${this.hostUrl}/api/phrase/update?p=${phrase.id}&text=${phrase.text}`
    console.log(url)
    return this.http.get(url).pipe(share());
  }

  updateGamePhrase(gid:string, phrase:Phrase){
    let url = `${this.hostUrl}/api/game/phrase/update?g=${gid}&p=${phrase.id}&text=${phrase.text}`
    return this.http.get(url).pipe(share());
  }

  addGameAdmin(gid:string, email:string){
    let headers = new HttpHeaders();
    headers.append('Content-Type', 'application/json');

    let options = { headers: headers };
    let url = `${this.hostUrl}/api/game/admin/add`

    let body = new FormData();
    body.append('g', gid);
    body.append('email', email);

    return this.http.post(url,body, options).pipe(share());
  }

  removeGameAdmin(gid:string, email:string){
    let url = `${this.hostUrl}/api/game/admin/remove?g=${gid}&email=${email}`
    return this.http.delete(url).pipe(share());
  }

  addAdmin(email:string){
    let headers = new HttpHeaders();
    headers.append('Content-Type', 'application/json');

    let options = { headers: headers };
    let url = `${this.hostUrl}/api/admin/add`

    let body = new FormData();
    body.append('email', email);

    return this.http.post(url,body, options).pipe(share());
  }

  removeAdmin(email:string){
    let url = `${this.hostUrl}/api/admin/remove?email=${email}`
    return this.http.delete(url).pipe(share());
  }

  getAdmins(){
    let url = `${this.hostUrl}/api/admin/list`
    return this.http.get(url).pipe(share());
  }

  messageReceived(gid:string, mid:string){
    let headers = new HttpHeaders();
    headers.append('Content-Type', 'application/json');
    let options = { headers: headers };
    let body = new FormData();
    body.append('g', gid);
    body.append('m', mid);
    let url = `${this.hostUrl}/api/message/receive`
    return this.http.post(url, body).pipe(share());
  }

}
