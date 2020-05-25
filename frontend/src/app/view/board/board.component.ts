import { Component, OnInit, ViewChild } from '@angular/core';
import { Observable, of as observableOf } from 'rxjs';
import { DataService, Phrase} from '../../service/data.service'
import {AuthService, Player} from '../../service/auth.service'
import {GameService, Board, Message, Game} from '../../service/game.service'
import { Router, ActivatedRoute } from '@angular/router';
import {ItemComponent} from './item/item.component'



@Component({
  selector: 'app-board',
  templateUrl: './board.component.html',
  styleUrls: ['./board.component.scss']
})
export class BoardComponent implements OnInit {

  @ViewChild(ItemComponent)
  private itemComponent: ItemComponent;
  public itemComponents: ItemComponent[] = [];

  public id:string;
  public board: Observable<any>;
  public phrases: Observable<any[]>;
  public currentState:any = {};
  public player:Player;
  public boardid:string;
  public messages: Observable<any[]>;
  public bingo:boolean=false;
  public game:Observable<any>;

  constructor(public data:DataService, public auth:AuthService, public gameService:GameService, public router:Router, route: ActivatedRoute,) {
    if (!auth.isAuth()){
      auth.logout("not authed")
    }
    this.id = route.snapshot.paramMap.get('id');
    this.player = auth.getPlayer(); 
    
    if (this.player.email == "undefined"){
      auth.logout("not authed")
    }
    
    this.board = gameService.getBoard(this.player.name, this.id);
    this.board.subscribe(val=>{this.boardid=val.id; if (val.bingodeclared){this.declareBingo()}})
    gameService.getGame(this.id).subscribe(val=>{let g:Game = val as Game; this.game=observableOf(g)});
   }

  ngOnInit(): void {
    this.messages = this.data.getMessages(this.id);
    this.messages.subscribe(ms=>{this.listenForBingo(ms),this.listenForReset(ms)})
  }

  ngOnChanges():void{
      if (this.bingo){
        this.declareBingo();
      }
  }


  declareBingo(){
    this.bingo=true;
      console.log("Bingo Declared");
      this.showBingo()
      this.itemComponents.forEach(function(child){
        child.disable();
      })
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

    if (msg.operation == "reset"){
      let ignoreid = localStorage.getItem(msg.id);
      if (ignoreid == null){
        this.auth.logout("reset message received for " + this.boardid, msg.id )
      }
    }
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
    
  }





}
