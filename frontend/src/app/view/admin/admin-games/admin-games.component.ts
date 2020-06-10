import { Component, OnInit } from '@angular/core';
import { GameService, Game } from 'src/app/service/game.service';
import { BehaviorSubject, Observable, of as observableOf  } from 'rxjs';

@Component({
  selector: 'app-admin-games',
  templateUrl: './admin-games.component.html',
  styleUrls: ['./admin-games.component.scss']
})
export class AdminGamesComponent implements OnInit {
  public games:BehaviorSubject<Game[]> = new BehaviorSubject([]);
  constructor(public game:GameService) { 
    this.refreshGame();
  }

  ngOnInit(): void {
  }

  deactivateGame($event){
    let id = $event.target.id;
    $event.target.parentElement.parentElement.style.display = 'none';
    this.game.deactivateGame(id).subscribe(val=>{this.refreshGame();});
  }

  refreshGame(){
    this.game.getGames().subscribe(val=>{
      let games:Game[] = val as Game[];
      games.sort((a, b) => (a.created > b.created) ? 1 : -1)
      this.games.next(games); 
    } );
  }

}
