import { Component, OnInit } from '@angular/core';
import { GameService } from 'src/app/service/game.service';

@Component({
  selector: 'app-gamepicker',
  templateUrl: './gamepicker.component.html',
  styleUrls: ['./gamepicker.component.scss']
})
export class GamepickerComponent implements OnInit {
  public games:any;
  constructor(public game:GameService) { 
    this.game.getGamesForPlayer().subscribe(val=>{this.games=val} );
  }
  

  ngOnInit(): void {
  }

}
