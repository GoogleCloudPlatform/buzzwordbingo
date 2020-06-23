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

import { Component, OnInit, Input,ViewChild } from '@angular/core';
import { GameService, Game } from 'src/app/service/game.service';
import { BehaviorSubject, Observable, of as observableOf  } from 'rxjs';
import { Player} from 'src/app/service/auth.service'
import { ProgressspinnerComponent } from '../../widgets/progressspinner/progressspinner.component';
import { FormalertComponent } from '../../widgets/formalert/formalert.component';


@Component({
  selector: 'app-admin-admins',
  templateUrl: './admin-admins.component.html',
  styleUrls: ['./admin-admins.component.scss']
})
export class AdminAdminsComponent implements OnInit {
  @ViewChild(ProgressspinnerComponent ) spinner: ProgressspinnerComponent ;
  @ViewChild(FormalertComponent ) formalert: FormalertComponent ;
  public admins:BehaviorSubject<Player[]> = new BehaviorSubject([]);
  constructor(private gameService:GameService) { 
    this.refreshAdmins();
  }
  ngOnInit(): void {
  }

  refreshAdmins(){
    this.gameService.getAdmins().subscribe(val=>{
        let p:Player[] = val as Player[]; 
        this.admins.next(p);
        this.spinner.toggle();
    });
  }

  onAdminAdd(email:string){
    this.gameService.addAdmin(email).subscribe(val=>{this.refreshAdmins()});
    this.formalert.alert(`Added ${email} to the list of admins`);
  }

  onAdminRemove($event, email:string ){
    $event.target.parentElement.style.display = 'none';
    this.gameService.removeAdmin(email).subscribe(val=>{$event.target.parentElement.style.display = 'none'; this.refreshAdmins();});
    this.formalert.alert(`Removed ${email} from the list of admins`);
  }

}
