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

import { Component, OnInit, Input } from '@angular/core';
import {GameService, Record, Game} from '../../../../service/game.service';
@Component({
  selector: 'app-manage-phrase-form',
  templateUrl: './manage-phrase-form.component.html',
  styleUrls: ['./manage-phrase-form.component.scss']
})
export class ManagePhraseFormComponent implements OnInit {
  @Input() record:Record;
  @Input() game:Game;
  timeout = null;
  constructor(private gameService:GameService) { }

  ngOnInit(): void {
  }

  keyup(){
    let self = this;
    clearTimeout(this.timeout);
    this.timeout = setTimeout(function () {self.onPhraseSubmit()}, 1000);
  }

  onPhraseSubmit(){
    clearTimeout(this.timeout);
    console.log("Save Phrase")
    this.gameService.updateGamePhrase(this.game.id, this.record.phrase).subscribe();
  }

}
