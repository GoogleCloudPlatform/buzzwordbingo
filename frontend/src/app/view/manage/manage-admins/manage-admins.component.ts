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

import { Component, OnInit, Input, ViewChild } from '@angular/core';
import { GameService, Game } from 'src/app/service/game.service';
import { BehaviorSubject, Observable, of as observableOf  } from 'rxjs';
import { Router, ActivatedRoute } from '@angular/router';
import { ProgressspinnerComponent } from '../../widgets/progressspinner/progressspinner.component';
import { FormalertComponent } from '../../widgets/formalert/formalert.component';

@Component({
  selector: 'app-manage-admins',
  templateUrl: './manage-admins.component.html',
  styleUrls: ['./manage-admins.component.scss']
})
export class ManageAdminsComponent implements OnInit {
  @ViewChild(ProgressspinnerComponent ) spinner: ProgressspinnerComponent ;
  @ViewChild(FormalertComponent ) formalert: FormalertComponent ;
  public id:string;
  public game:BehaviorSubject<Game> = new BehaviorSubject(new Game);
  public gameid:string;
  constructor(private gameService:GameService, public router:Router, route: ActivatedRoute,) { 
    this.id = route.snapshot.paramMap.get('id');
    this.refreshGame();
  }

  refreshGame(){
    this.gameService.getGame(this.id).subscribe(val=>{
      let g:Game = val as Game; 
      this.gameid=g.id; 
      this.game.next(g);
      this.spinner.toggle();
    });
  }

  ngOnInit(): void {
  }

  onAdminAdd(email:string){
    this.gameService.addGameAdmin(this.gameid, email).subscribe(val=>{this.refreshGame()});
    this.formalert.alert(`Added ${email} to the list of admins`);
    this.refreshGame();
  }

  onAdminRemove($event, email:string){
    console.log($event)
    $event.target.parentElement.style.display = 'none';
    this.gameService.removeGameAdmin(this.gameid, email).subscribe(val=>{$event.target.parentElement.style.display = 'none'; this.refreshGame();});
    this.formalert.alert(`Removed ${email} from the list of admins`);
    
    
  }

}
