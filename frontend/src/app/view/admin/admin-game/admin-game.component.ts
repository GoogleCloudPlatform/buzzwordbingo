import { Component, OnInit } from '@angular/core';
import { GameService} from '../../../service/game.service'
import { AuthService} from '../../../service/auth.service'

@Component({
  selector: 'app-admin-game',
  templateUrl: './admin-game.component.html',
  styleUrls: ['./admin-game.component.scss']
})
export class AdminGameComponent implements OnInit {

  constructor(public game: GameService,public auth: AuthService) { }

  ngOnInit(): void {
  }

  createNewGame(){
    let name:HTMLInputElement = document.querySelector(".name") as HTMLInputElement;
    this.game.newGame(name.value)
    this.auth.logout("game was reset")
  }

}
