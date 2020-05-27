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
    this.game.getGamesForPlayer().subscribe(val=>{
      let games:Game[] = val as Game[];
      this.games=games; 
      if (games.length == 1){
        this.router.navigateByUrl('/game/' + val[0].id);
        return;
      }
    } );
  }
  

  ngOnInit(): void {
  }

}
