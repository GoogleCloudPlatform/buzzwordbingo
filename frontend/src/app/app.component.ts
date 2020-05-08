import { Component } from '@angular/core';
import {GameService} from '../app/service/game.service'

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  title = 'bingomeeting';

  constructor(private game: GameService) { 
    game.getActiveGame()
  }
}
