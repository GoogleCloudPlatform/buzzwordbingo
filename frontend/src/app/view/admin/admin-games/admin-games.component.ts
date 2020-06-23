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

@Component({
  selector: 'app-admin-games',
  templateUrl: './admin-games.component.html',
  styleUrls: ['./admin-games.component.scss']
})
export class AdminGamesComponent implements OnInit {
  @ViewChild(ProgressbarComponent ) bar: ProgressbarComponent ; 
  public games:BehaviorSubject<Game[]> = new BehaviorSubject([]);
  
  constructor(public game:GameService) { 
    this.refreshGame();
  }

  ngOnInit(): void {
  }

  deactivateGame($event){
    let id = $event.target.id;
    $event.target.parentElement.parentElement.style.display = 'none';
    this.game.deactivateGame(id).subscribe(val=>{this.refreshGame();});
  }

  refreshGame(){
    this.game.getGames().subscribe(val=>{
      let games:Game[] = val as Game[];
      games.sort((a, b) => (a.created > b.created) ? 1 : -1)
      this.games.next(games); 
      this.bar.toggle();
    } );
  }

}
