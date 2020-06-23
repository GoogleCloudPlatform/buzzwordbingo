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
import { Observable, of } from 'rxjs';
import { DataService, Phrase} from '../../../service/data.service'
import {GameService, Board, Message, Record} from '../../../service/game.service'
import {ProgressbarComponent} from 'src/app/view/widgets/progressbar/progressbar.component'
import { FormalertComponent } from '../../widgets/formalert/formalert.component';

@Component({
  selector: 'app-manage-boards',
  templateUrl: './manage-boards.component.html',
  styleUrls: ['./manage-boards.component.scss']
})
export class ManageBoardsComponent implements OnInit {
  @ViewChild(ProgressbarComponent ) bar: ProgressbarComponent ; 
  @ViewChild(FormalertComponent ) formalert: FormalertComponent ;
  @Input() gid:string;
  public boards: Observable<any[]>;
  constructor(public data:DataService, public gameService:GameService) { }

  ngOnInit(): void {
    this.boards = this.data.getBoards(this.gid)
    this.boards.subscribe(val=>{this.bar.toggle()});
  }

  reset(bid:string, gid:string){
    this.gameService.resetboard(bid, gid);
    this.formalert.alert(`Board reset`);
  }

  onAdminAdd(email:string){
    this.gameService.addGameAdmin(this.gid, email).subscribe();
  }


}
