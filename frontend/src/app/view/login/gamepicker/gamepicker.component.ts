import { Component, OnInit,ViewChild } from '@angular/core';
import { GameService, Game } from 'src/app/service/game.service';
import {Router, ActivatedRoute} from '@angular/router';
import { BehaviorSubject, Observable, of as observableOf  } from 'rxjs';
import {ProgressbarComponent} from 'src/app/view/widgets/progressbar/progressbar.component'

@Component({
  selector: 'app-gamepicker',
  templateUrl: './gamepicker.component.html',
  styleUrls: ['./gamepicker.component.scss']
})
export class GamepickerComponent implements OnInit {
  @ViewChild(ProgressbarComponent ) bar: ProgressbarComponent ; 
  public games:BehaviorSubject<Game[]> = new BehaviorSubject([]);
  
  constructor(public game:GameService, public router:Router) { 
    this.refreshGames();
  }

  refreshGames(){
    this.game.GetGamesForKey().subscribe(val=>{
      let games:Game[] = val as Game[];
      games.sort((a, b) => (a.created > b.created) ? 1 : -1)
      this.games.next(games); 
      this.bar.toggle();
    } );
  }
  

  ngOnInit(): void {
  }

}
