import { Component, OnInit, Input } from '@angular/core';
import { Observable, of as observableOf } from 'rxjs';
import {GameService, Board, Message, Game} from '../../../service/game.service'
import { Router, ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-manage-phrases',
  templateUrl: './manage-phrases.component.html',
  styleUrls: ['./manage-phrases.component.scss']
})
export class ManagePhrasesComponent implements OnInit {
  public game:Observable<Game>;
  public id:string;
  constructor(private gameService:GameService, public router:Router, route: ActivatedRoute) { 
    this.id = route.snapshot.paramMap.get('id');
    this.gameService.getGame(this.id).subscribe(val=>{let g:Game = val as Game; this.game=observableOf(g);});
  }

  ngOnInit(): void {
  }

}
