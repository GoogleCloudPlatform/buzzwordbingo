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

  constructor(public data:DataService, public auth:AuthService, public game:GameService, router:Router) {
    let self = this;
    if (!auth.isAuth()){
      router.navigateByUrl('/login');
    }

    this.player = auth.getPlayer(); 
    
    
    
    if (this.player.email != "undefined"){
      this.board = game.getBoard(this.player.email, this.player.name);
    }
    
    this.board.subscribe(val=>{this.boardid=val.id; console.log(val); if (val.bingodeclared){this.declareBingo()}})

   }

  ngOnInit(): void {
    this.messages = this.data.getMessages(this.game.game.id);
    this.messages.subscribe(ms=>{this.listenForBingo(ms)})
  }

  ngAfterViewInit() {
    if (this.bingo){
      console.log("Bingo declared in afterviewinit")
      this.itemComponent.disable();
    }
  }
  ngOnChanges(){
    console.log("On Changes called");
  }

  declareBingo(){
    this.bingo=true;
      console.log("Bingo Declared");
      console.log(this.itemComponents);
      this.itemComponents.forEach(function(child){
        console.log("Looping through children")
        child.disable();
      })
  }

  listenForBingo(messages:Message[]){
    let self = this;
    console.log(messages)
    let msg:Message = messages[messages.length-1] as Message;
    console.log(msg)
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

    if (this.checkBingo()){
      alert("BINGO!")
    }
    
  }

  receiveChild($event) {
    let child = $event;
    this.itemComponents.push(child)
    
  }


  checkBingo(){
    let counts = {};
    let diag1 = ["b1", "i2", "n3", "g4", "o5"];
    let diag2 = ["b5", "i4", "n3", "g2", "o1"];

    let keys = Object.values(this.currentState) as Phrase[];

    keys.forEach(function(phrase) {
      var column = phrase.tid.charAt(0);
      var row= phrase.tid.charAt(1);
      if (phrase.selected){
          counts[column] = (counts[column] || 0) + 1;
          counts[row] = (counts[row] || 0) + 1;

          if (diag1.indexOf(phrase.tid) >= 0) {
              counts["diag1"] = (counts["diag1"] || 0) + 1;
          }

          if (diag2.indexOf(phrase.tid) >= 0) {
              counts["diag2"] = (counts["diag2"] || 0) + 1;
          }
      }
    });
    for (let key in counts) {
      if (counts[key] == 5){
          return true;
      } 
    }
    return false;


  }



}
