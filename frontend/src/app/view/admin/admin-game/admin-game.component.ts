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

  constructor(public game: GameService,public auth: AuthService, private router: Router) { }

  ngOnInit(): void {
  }

  createNewGame(name:string){
    console.log(name);
    this.game.newGame(name).subscribe(val => {
      let g:Game = val as Game;
      this.router.navigateByUrl('/game/' + g.id);
    })
    ;
  }

}
