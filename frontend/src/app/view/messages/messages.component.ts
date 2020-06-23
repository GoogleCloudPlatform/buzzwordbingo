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

import { Component, OnInit, Input, ChangeDetectorRef, ViewChild} from '@angular/core';
import {GameService, Message} from '../../service/game.service'
import { Observable } from 'rxjs';
import { ProgressspinnerComponent } from '../widgets/progressspinner/progressspinner.component';


@Component({
  selector: 'app-messages',
  templateUrl: './messages.component.html',
  styleUrls: ['./messages.component.scss']
})
export class MessagesComponent implements OnInit {
  @ViewChild(ProgressspinnerComponent ) spinner: ProgressspinnerComponent ; 
  @Input() messages: Observable<any>;
  @Input() target:string;
  

  constructor(private cdref: ChangeDetectorRef) { }

  ngOnInit(): void {
    this.scrollDownWindow();
    this.messages.subscribe(
      val=>{
        this.scrollDownWindow();
        this.spinner.toggle();
      })
  }

  ngAfterViewInit(): void {
  }
  ngOnChanges(): void {
    this.scrollDownWindow();
  }

  findInAudience(message:Message){
    
    let self = this;
    let result:boolean = false;
    message.audience.forEach(function(val){
        if ((val == self.target) || (val == "all")) {
          result = true
        } 
    })
    return result;    
  }


  scrollDownWindow(){
    let d = document.querySelector('.messages');
    if (d) {
      d.scrollTop = d.scrollHeight;
    }
    this.cdref.detectChanges();
  }


}
