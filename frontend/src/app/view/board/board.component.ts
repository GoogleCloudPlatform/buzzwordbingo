import { Component, OnInit, ViewChild } from '@angular/core';
import { Observable, of } from 'rxjs';
import { DataService, Phrase} from '../../service/data.service'
import {AuthService, Player} from '../../service/auth.service'
import {GameService, Board, Message} from '../../service/game.service'
import {Router} from '@angular/router';
import {ItemComponent} from '../item/item.component'



@Component({
  selector: 'app-board',
  templateUrl: './board.component.html',
  styleUrls: ['./board.component.scss']
})
export class BoardComponent implements OnInit {

  @ViewChild(ItemComponent)
  private itemComponent: ItemComponent;
  public itemComponents: ItemComponent[] = [];

  public board: Observable<any>;
  public phrases: Observable<any[]>;
  public currentState:any = {};
  public player:Player;
  public boardid:string;
  public messages: Observable<any[]>;
  public bingo:boolean=false;

  constructor(public data:DataService, public auth:AuthService, public game:GameService, public router:Router) {
    let self = this;
    if (!auth.isAuth()){
      localStorage.clear();
      router.navigateByUrl('/login');
    }

    this.player = auth.getPlayer(); 
    
    
    
    if (this.player.email != "undefined"){
      this.board = game.getBoard(this.player.email, this.player.name);
    }
    
    this.board.subscribe(val=>{this.boardid=val.id; if (val.bingodeclared){this.declareBingo()}})

   }

  ngOnInit(): void {
    this.messages = this.data.getMessages(this.game.game.id);
    this.messages.subscribe(ms=>{this.listenForBingo(ms),this.listenForReset(ms)})
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
      this.router.navigateByUrl('/login');
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
