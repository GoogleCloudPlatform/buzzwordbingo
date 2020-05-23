import { Component, OnInit } from '@angular/core';
import { Observable, of  } from 'rxjs';
import {Router} from '@angular/router';
import {AuthService, Player} from '../../service/auth.service'
import { GameService, Game } from 'src/app/service/game.service';


@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent implements OnInit {

  public identity:Observable<any>;
  public games:any;

  constructor(private auth:AuthService, private router: Router, public game:GameService) { 
    this.identity =auth.identifyPlayer();
    this.game.getGamesForPlayer().subscribe(val=>{this.games=val} );
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
    let gameid = "";
    this.games.forEach(v => { 
      if (v.active){
        gameid = v.id;
        return
      }
    });
      
    

    this.router.navigateByUrl('/game/' + gameid);

  }

}
