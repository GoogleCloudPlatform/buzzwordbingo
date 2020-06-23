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


@Component({
  selector: 'app-manage-master',
  templateUrl: './manage-master.component.html',
  styleUrls: ['./manage-master.component.scss']
})
export class ManageMasterComponent implements OnInit {
  @ViewChild(ProgressbarComponent ) bar: ProgressbarComponent ; 
  @Input() id:string;
  public records: Observable<any[]>;
  constructor(public data:DataService, public game:GameService) { }

  ngOnInit(): void {
     this.records = this.data.getRecords(this.id)
     this.records.subscribe(val=>{this.bar.toggle()});
  }

  playerCount(record:Record):number{
    if (record.players != null){
      return record.players.length
    }
    return 0;
  }

  isActive(record:Record):boolean{
    if (record.players != null && record.players.length > 0){
      return true
    }
    return false;
  }

}
