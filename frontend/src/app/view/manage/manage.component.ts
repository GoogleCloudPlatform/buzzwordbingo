import { Component, OnInit } from '@angular/core';
import {GameService, Board, Message, Game} from '../../service/game.service';
import { DataService, Phrase} from '../../service/data.service'
import { Observable, of as observableOf } from 'rxjs';
import { Router, ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-manage',
  templateUrl: './manage.component.html',
  styleUrls: ['./manage.component.scss']
})
export class ManageComponent implements OnInit {
  public id:string;
  public messages: Observable<any[]>;
  public game:Observable<Game>;
  public inviteLink:string;
  message_target:string = "admin";

  constructor(public data:DataService, public gameService:GameService, public router:Router, route: ActivatedRoute,) {
    this.id = route.snapshot.paramMap.get('id');
    gameService.getGame(this.id).subscribe(val=>{let g:Game = val as Game; this.game=observableOf(g)});
    this.inviteLink = "http://" + window.location.hostname + "/invite";
  }

  ngOnInit(): void {
    this.messages = this.data.getMessages(this.id);
    this.messages.subscribe()
  }

  deactivateGame(){
    this.gameService.deactivateGame(this.id).subscribe();
  }



}
