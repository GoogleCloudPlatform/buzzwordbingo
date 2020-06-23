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
import { GameService, Game} from '../../../service/game.service'
import { AuthService} from '../../../service/auth.service'
import { Router } from '@angular/router';

@Component({
  selector: 'app-admin-game',
  templateUrl: './admin-game.component.html',
  styleUrls: ['./admin-game.component.scss']
})
export class AdminGameComponent implements OnInit {

  gameName:string="";

  constructor(private game: GameService, private router: Router, private auth: AuthService) { }

  ngOnInit(): void {
  }

  createNewGame(name:string){
    let player = this.auth.getPlayer()


    this.game.newGame(name,player.name).subscribe(val => {
      let g:Game = val as Game;
      this.router.navigateByUrl('/game/' + g.id);
    })
    ;
  }

}
