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
