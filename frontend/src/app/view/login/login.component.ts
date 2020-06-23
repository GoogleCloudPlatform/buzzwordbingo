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
import { Observable, of, Subscription, BehaviorSubject  } from 'rxjs';
import {Router, ActivatedRoute} from '@angular/router';
import {AuthService, Player} from '../../service/auth.service'
import { GameService, Game } from 'src/app/service/game.service';
import { DataService } from 'src/app/service/data.service';
import { GoogleAuthService } from 'src/app/service/googleauth.service';


@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent implements OnInit {

  public id:string="";
  public identity:Observable<any>;
  public games:any;
  public ga:BehaviorSubject<boolean> = new BehaviorSubject<boolean>(false);

  constructor(private auth:AuthService, 
              public router:Router, 
              route: ActivatedRoute, 
              public game:GameService, 
              public dataService:DataService,
              public googleAuth:GoogleAuthService) { 
    this.id = route.snapshot.paramMap.get('id');
    this.identity =auth.identifyPlayer();
    this.game.GetGamesForKey().subscribe(val=>{this.games=val; } );
    googleAuth.authed.subscribe(val=>{
        this.ga.next(val);
        if (val){
          let login = document.querySelector(".login") as HTMLElement;
          let message = document.querySelector(".message") as HTMLElement;
          if (login != null && message != null){
            login.style.display="block";
            message.style.display="none";
          }
         
        }
    });
  }

  ngOnInit(): void {
  }

  submitPlayer(){

    let nameInput:HTMLInputElement = document.querySelector("#name") as HTMLInputElement;
    let emailInput:HTMLInputElement = document.querySelector("#email") as HTMLInputElement;
    if (nameInput.value == ""){
      let alert:HTMLInputElement = document.querySelector(".alert") as HTMLInputElement;
      alert.style.display = "block";
      return;
    }

    this.auth.setPlayer(nameInput.value, emailInput.value);
    let gameids = [];
    this.games.forEach(v => { 
      if (v.active){
        gameids.push(v.id);
        return
      }
    });

    if (this.id != null){
      this.router.navigateByUrl('/game/' + this.id);
      return;
    }
      
    this.router.navigateByUrl('/gamepicker');
    return;
    

  }


}
