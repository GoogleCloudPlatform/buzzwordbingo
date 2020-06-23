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
import { Observable, of } from 'rxjs';
import { DataService, Phrase} from '../../../service/data.service'
import {GameService, Board, Message, Record} from '../../../service/game.service'
import { ProgressspinnerComponent } from '../../widgets/progressspinner/progressspinner.component';



@Component({
  selector: 'app-admin-phrases',
  templateUrl: './admin-phrases.component.html',
  styleUrls: ['./admin-phrases.component.scss']
})
export class AdminPhrasesComponent implements OnInit {
  @ViewChild(ProgressspinnerComponent ) spinner: ProgressspinnerComponent ;
  public phrases: Observable<any[]>;
  
  constructor(public data:DataService, public game:GameService) { }

  ngOnInit(): void {

    this.phrases = this.data.getPhrases()
    this.phrases.subscribe(val=>{this.spinner.toggle()});
  }

  

}
