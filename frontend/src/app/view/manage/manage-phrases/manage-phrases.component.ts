import { Component, OnInit, Input, ViewChild } from '@angular/core';
import { Observable, of as observableOf } from 'rxjs';
import {GameService, Board, Message, Game} from '../../../service/game.service'
import { Router, ActivatedRoute } from '@angular/router';
import { ProgressspinnerComponent } from '../../widgets/progressspinner/progressspinner.component';


@Component({
  selector: 'app-manage-phrases',
  templateUrl: './manage-phrases.component.html',
  styleUrls: ['./manage-phrases.component.scss']
})
export class ManagePhrasesComponent implements OnInit {
  @ViewChild(ProgressspinnerComponent ) spinner: ProgressspinnerComponent ;
  public game:Observable<Game>;
  public id:string;
  constructor(private gameService:GameService, public router:Router, route: ActivatedRoute) { 
    this.id = route.snapshot.paramMap.get('id');
    this.gameService.getGame(this.id).subscribe(val=>{
        let g:Game = val as Game; 
        this.game=observableOf(g);
        this.spinner.toggle();
    });
  }

  ngOnInit(): void {
  }

}
