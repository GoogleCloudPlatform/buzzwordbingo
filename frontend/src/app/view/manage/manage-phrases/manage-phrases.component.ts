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
import { Observable, of as observableOf } from 'rxjs';
import {GameService, Board, Message, Game} from '../../../service/game.service'
import { Router, ActivatedRoute } from '@angular/router';
import { ProgressspinnerComponent } from '../../widgets/progressspinner/progressspinner.component';


@Component({
  selector: 'app-manage-phrases',
  templateUrl: './manage-phrases.component.html',
  styleUrls: ['./manage-phrases.component.scss']
})
export class ManagePhrasesComponent implements OnInit {
  @ViewChild(ProgressspinnerComponent ) spinner: ProgressspinnerComponent ;
  public game:Observable<Game>;
  public id:string;
  constructor(private gameService:GameService, public router:Router, route: ActivatedRoute) { 
    this.id = route.snapshot.paramMap.get('id');
    this.gameService.getGame(this.id).subscribe(val=>{
        let g:Game = val as Game; 
        this.game=observableOf(g);
        this.spinner.toggle();
    });
  }

  ngOnInit(): void {
  }

}
