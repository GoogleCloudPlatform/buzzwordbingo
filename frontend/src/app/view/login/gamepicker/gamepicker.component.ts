import { Component, OnInit } from '@angular/core';
import { GameService, Game } from 'src/app/service/game.service';
import {Router, ActivatedRoute} from '@angular/router';

@Component({
  selector: 'app-gamepicker',
  templateUrl: './gamepicker.component.html',
  styleUrls: ['./gamepicker.component.scss']
})
export class GamepickerComponent implements OnInit {
  public games:any;
  constructor(public game:GameService, public router:Router) { 
    this.game.GetGamesForKey().subscribe(val=>{
      let games:Game[] = val as Game[];
      games.sort((a, b) => (a.created > b.created) ? 1 : -1)
      this.games=games; 
    } );
  }
  

  ngOnInit(): void {
  }

}
