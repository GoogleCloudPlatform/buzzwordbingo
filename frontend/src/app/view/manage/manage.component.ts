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
import {GameService, Board, Message, Game} from '../../service/game.service';
import { DataService, Phrase} from '../../service/data.service'
import {AuthService, Player} from 'src/app/service/auth.service'
import { Observable, of as observableOf } from 'rxjs';
import { Router, ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-manage',
  templateUrl: './manage.component.html',
  styleUrls: ['./manage.component.scss']
})
export class ManageComponent implements OnInit {
  public id:string;
  public messages: Observable<any[]>;
  public game:Observable<Game>;
  public inviteLink:string;
  message_target:string = "admin";

  constructor(public data:DataService, public gameService:GameService, public router:Router, route: ActivatedRoute, public auth:AuthService) {
    if (!auth.isAuth()){
      auth.logout("not authed")
    }
    
    
    this.id = route.snapshot.paramMap.get('id');
    gameService.getGame(this.id).subscribe(val=>{let g:Game = val as Game; this.game=observableOf(g)});
    this.inviteLink = "http://" + window.location.hostname + "/invite";
  }

  ngOnInit(): void {
    this.messages = this.data.getMessagesAdmin(this.id);
    this.messages.subscribe()
  }

  deactivateGame(){
    this.gameService.deactivateGame(this.id).subscribe();
  }



}
