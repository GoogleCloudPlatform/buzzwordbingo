import { Component, OnInit, ViewChild } from '@angular/core';
import { Observable, of as observableOf } from 'rxjs';
import { DataService, Phrase} from '../../service/data.service'
import {AuthService, Player} from '../../service/auth.service'
import {GameService, Board, Message, Game} from '../../service/game.service'
import { Router, ActivatedRoute } from '@angular/router';
import {ItemComponent} from './item/item.component'
import { map, share } from 'rxjs/operators';



@Component({
  selector: 'app-board',
  templateUrl: './board.component.html',
  styleUrls: ['./board.component.scss']
})
export class BoardComponent implements OnInit {

  @ViewChild(ItemComponent)
  private itemComponent: ItemComponent;
  public itemComponents: ItemComponent[] = [];

  public gid:string;
  public board: Observable<any>;
  public phrases: Observable<any[]>;
  public currentState:any = {};
  public player:Player;
  public boardid:string;
  public messages: Observable<any[]>;
  public bingo:boolean=false;
  public game:Observable<any>;
  public inviteLink:string;

  constructor(public data:DataService, public auth:AuthService, public gameService:GameService, public router:Router, route: ActivatedRoute,) {
    if (!auth.isAuth()){
      auth.logout("not authed")
    }
    this.inviteLink = "http://" + window.location.hostname + "/invite";
    this.gid = route.snapshot.paramMap.get('id');
    this.player = auth.getPlayer(); 
    
    if (this.player.email == "undefined"){
      auth.logout("not authed")
    }

    let block = false;
    if (!block){
    
      this.board = gameService.getBoard(this.player.name, this.gid).pipe(share());
    
      this.board.subscribe(val=>{
        block = true;
        this.boardid=val.id; 
        this.phrases = data.getGameBoard(this.gid, this.boardid).pipe(map(val => {
          let phrases:Phrase[] = val as Phrase[]
          phrases= phrases.sort((a, b) => (a.displayorder > b.displayorder) ? 1 : -1)
          return phrases;
        }))
        if (val.bingodeclared){
          this.declareBingo()
        }
        block = false;
      })
    }
   
    
    


    gameService.getGame(this.gid).subscribe(val=>{let g:Game = val as Game; this.game=observableOf(g)});
   }

  ngOnInit(): void {
    let player:Player = this.auth.getPlayer();
    this.messages = this.data.getMessages(this.gid, player.email);
    this.messages.subscribe(ms=>{this.listenForBingo(ms),this.listenForReset(ms)})
  }

  ngOnChanges():void{
      if (this.bingo){
        this.declareBingo();
      }
  }

  declareBingo(){
    this.bingo=true;
      this.showBingo();
      this.itemComponents.forEach(function(child){
        child.disable();
      })
  }
  reset(bid:string, gid:string){
    this.gameService.resetboard(bid, gid);
  }


  showBingo(){
    let board = document.querySelector(".header-container");
    board.classList.add("header-bingo");
  }

  listenForReset(messages:Message[]){
    let self = this;
    let msg:Message = messages[messages.length-1] as Message;
    if (!msg || typeof msg == "undefined"){
      return;
    }
    let halt:boolean = true;
    msg.audience.forEach(function(aud){
      if( (aud == self.auth.getPlayer().email) ){
        halt = false;
      }
      if( (aud == "all") ){
        halt = false;
      }
    })

    if (halt){
      return;
    }

    if (msg.operation == "reset" && !msg.received){
      console.log(msg);
      this.gameService.messageReceived(this.gid, msg.id).subscribe(val=>{
        this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
          this.router.navigateByUrl('/game/'+this.gid);
        }); 
      });
      
      
      return;
    }
    return;
  }



  listenForBingo(messages:Message[]){
    let self = this;
    let msg:Message = messages[messages.length-1] as Message;
    if (!msg || typeof msg == "undefined"){
      return;
    }
    let halt:boolean = true;
    msg.audience.forEach(function(aud){
      if( (aud == self.auth.getPlayer().email) ){
        halt = false;
      }
    })

    if (halt){
      return;
    }

    if (msg.bingo){
      this.declareBingo()
    }
  }

  recievePhrase($event) {
    let phrase = $event;

    if (phrase.selected){
      this.currentState[phrase.id] = phrase;
    } else {
      delete this.currentState[phrase.id];
    }

    
  }

  receiveChild($event) {
    let child = $event;
    this.itemComponents.push(child)

    if(this.itemComponents.length == 25 && this.bingo){
      this.itemComponents.forEach(function(child){
        child.disable();
      })
    }
  }





}
