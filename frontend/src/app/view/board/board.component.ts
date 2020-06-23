/**
 * Copyright 2020 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { Component, OnInit, ViewChild, ViewChildren } from '@angular/core';
import { Observable, of as observableOf, Subscription } from 'rxjs';
import { DataService, Phrase} from '../../service/data.service'
import {AuthService, Player} from '../../service/auth.service'
import {GameService, Board, Message, Game} from '../../service/game.service'
import { Router, ActivatedRoute } from '@angular/router';
import {ItemComponent} from './item/item.component'
import { map, share, debounceTime } from 'rxjs/operators';
import { SnackbarComponent } from '../widgets/snackbar/snackbar.component';



@Component({
  selector: 'app-board',
  templateUrl: './board.component.html',
  styleUrls: ['./board.component.scss']
})
export class BoardComponent implements OnInit {
  // @ViewChildren(ItemComponent)
  @ViewChild('snackbar') sb: SnackbarComponent;
 
  public itemComponents: ItemComponent[] = [];

  public gid:string;
  public board: Observable<any>;
  public phrases: Observable<any[]>;
  public currentState:any = {"started":false};
  public player:Player;
  public boardid:string;
  public messages: Observable<any[]>;
  public bingo:boolean=false;
  public game:Observable<any>;
  public inviteLink:string;
  public showInvitelink:boolean = false;
  
  private messageSubscription:Subscription;
  private gameSubscription:Subscription;
  private boardSubscription:Subscription;

  constructor(public data:DataService, public auth:AuthService, public gameService:GameService, public router:Router, route: ActivatedRoute,) {
    if (!auth.isAuth()){
      auth.logout("not authed")
    }
    
    
    
    this.gid = route.snapshot.paramMap.get('id');
    this.player = auth.getPlayer(); 
    this.inviteLink = "http://" + window.location.hostname + "/invite/" + this.gid;
    
    if (this.player.email == "undefined"){
      auth.logout("not authed")
    }

    this.gameSubscription = gameService.getGame(this.gid).subscribe(val=>{
      let g:Game = val as Game; 
      this.game=observableOf(g);
      if (g.players.length == 1 && g.players[0].email == this.player.email){
        this.showInvitelink = true;
      }
    });
    this.messages = this.data.getMessages(this.gid, this.player.email);
    this.messageSubscription = this.messages.subscribe(ms=>{this.listenForBingo(ms);this.listenForReset(ms)})
    this.board = gameService.getBoard(this.player.name, this.gid).pipe(debounceTime(1000),share());

    this.handleBoard();

   
   }

   handleBoard(){
   
    let self = this;
    let block = false;
    if (!block){
    this.boardSubscription = this.board.subscribe(val=>{
      block = true;
      this.boardid = val.id; 
      this.phrases = this.data.getGameBoard(this.gid, this.boardid).pipe(map(val => {
        let phrases:Phrase[] = val as Phrase[];

        // This section is weird, but basically it allows an admin to revert a 
        // bingo by eliminating a selelcted square
        phrases.forEach(function(v){
          let old = self.currentState[v.id];
          if (old != null && old.text != v.text){
            self.gameService.getBoard(self.player.name, self.gid).subscribe(val=>{
              if (val.bingodeclared){
                self.declareBingo();
              } else{
                self.hideBingo();
                self.bingo = false;
                self.enableBoard();
              }
            });
          }
          self.currentState[v.id] = v;
        });
        if (!self.currentState.started){
          self.currentState.started = true;
        } 


        phrases = phrases.sort((a, b) => (a.displayorder > b.displayorder) ? 1 : -1)
        return phrases;
      }))
      if (val.bingodeclared){
        this.declareBingo()
      } 
      block = false;
    })

    }
   }

  ngOnInit(): void {
    
  }

  ngOnDestroy() {
    this.messageSubscription.unsubscribe();
    this.boardSubscription.unsubscribe();
    this.gameSubscription.unsubscribe();
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

  copyInviteLink(){
      navigator.clipboard.writeText(this.inviteLink);
  }

  hideInviteLink(){
    this.showInvitelink = false;
  }

  showBingo(){
    let board = document.querySelector(".header-container");
    board.classList.add("header-bingo");
    this.sb.show("You have bingo!");
  }

  hideBingo(){
    let board = document.querySelector(".header-container");
    board.classList.remove("header-bingo");
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
      
      this.gameService.messageReceived(this.gid, msg.id).pipe(debounceTime(1000)).subscribe(val=>{
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
    this.itemComponents.push(child);

    if(this.itemComponents.length == 25 && this.bingo){
      this.itemComponents.forEach(function(child){
        child.disable();
      })
    } else if (!this.bingo){
        this.hideBingo();
    }

    
  }

  enableBoard(){
    this.itemComponents.forEach(function(child){
      child.enable();
    })
  }





}
