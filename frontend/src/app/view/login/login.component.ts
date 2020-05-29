import { Component, OnInit } from '@angular/core';
import { Observable, of  } from 'rxjs';
import {Router, ActivatedRoute} from '@angular/router';
import {AuthService, Player} from '../../service/auth.service'
import { GameService, Game } from 'src/app/service/game.service';


@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent implements OnInit {

  public id:string="";
  public identity:Observable<any>;
  public games:any;

  constructor(private auth:AuthService, public router:Router, route: ActivatedRoute, public game:GameService) { 
    this.id = route.snapshot.paramMap.get('id');
    console.log(this.id);
    this.identity =auth.identifyPlayer();
    this.game.getGamesForPlayer().subscribe(val=>{this.games=val; } );
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
