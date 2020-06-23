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

import { Component, OnInit } from '@angular/core';
import {GameService, Board, Message} from '../../service/game.service';
import { DataService, Phrase} from '../../service/data.service'
import { Observable, of } from 'rxjs';
import { Router, ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-admin',
  templateUrl: './admin.component.html',
  styleUrls: ['./admin.component.scss']
})
export class AdminComponent implements OnInit {
  public id:string;
  public messages: Observable<any[]>;
  message_target:string = "admin";
  constructor(public data:DataService, public game:GameService, public router:Router, route: ActivatedRoute,) {
    this.id = route.snapshot.paramMap.get('id');
   }

  ngOnInit(): void {
  }

}
