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

import { Component, OnInit, ViewChild } from '@angular/core';
import { GameService, Game } from 'src/app/service/game.service';
import { BehaviorSubject, Observable, of as observableOf  } from 'rxjs';
import {ProgressbarComponent} from 'src/app/view/widgets/progressbar/progressbar.component'
import { FormalertComponent } from '../../widgets/formalert/formalert.component';

@Component({
  selector: 'app-admin-games',
  templateUrl: './admin-games.component.html',
  styleUrls: ['./admin-games.component.scss']
})
export class AdminGamesComponent implements OnInit {
  @ViewChild(ProgressbarComponent ) bar: ProgressbarComponent ; 
  @ViewChild(FormalertComponent ) formalert: FormalertComponent ;
  public games:BehaviorSubject<Game[]> = new BehaviorSubject([]);
  private limit:number = 5;
  private tokens:number[] = [];
  public showPrev = false;
  public showNext = false;
  constructor(public game:GameService) { 
    this.refreshGame();
  }

  ngOnInit(): void {
  }

  deactivateGame($event, gid:string){
    $event.target.parentElement.parentElement.style.display = 'none';
    this.formalert.alert(`Game has been deactivated`);
    this.game.deactivateGame(gid).subscribe(val=>{this.refreshGame();});
  }

  refreshGame(){
    let token = Math.round(new Date().getTime() /1000);
    this.game.getGames(this.limit, token.toString()).subscribe(val=>{
      let games:Game[] = val as Game[];
      games.sort((a, b) => (a.created < b.created) ? 1 : -1)
      this.games.next(games); 
      this.bar.toggle();
      this.showNext = true;
      this.tokens.push(token);
    } );
  }

  next(){


    let gs = this.games.getValue();
    let date = gs[gs.length -1].created;
    let t = new Date(date);
    let token = Math.round(t.getTime() /1000);
    this.bar.show();

    this.game.getGames(this.limit, token.toString()).subscribe(val=>{
      let games:Game[] = val as Game[];
      games.sort((a, b) => (a.created < b.created) ? 1 : -1)

      this.games.next(games); 
      this.bar.toggle();
      this.showPrev = true;
      

      if (this.tokens[this.tokens.length -1] == token){
        this.showNext = false;
      } else {
        this.tokens.push(token);
        this.showNext = true;
      }

      if (games.length != this.limit){
        this.showNext = false;
      }

    } );

  }

  prev(){

    this.tokens.pop();
    let token = this.tokens[this.tokens.length-1];
    this.bar.show();

    this.game.getGames(this.limit, token.toString()).subscribe(val=>{
      let games:Game[] = val as Game[];
      games.sort((a, b) => (a.created < b.created) ? 1 : -1)

      this.games.next(games); 
      this.bar.toggle();
      this.showNext = true;

      if (this.tokens.length == 1){
        this.showPrev = false;
      } else {
        this.showPrev = true;
      }
     
    } );



  }

}
