import { Component, OnInit } from '@angular/core';
import {GameService, Board, Message} from '../../service/game.service';
import { DataService, Phrase} from '../../service/data.service'
import { Observable, of } from 'rxjs';

@Component({
  selector: 'app-admin',
  templateUrl: './admin.component.html',
  styleUrls: ['./admin.component.scss']
})
export class AdminComponent implements OnInit {
  public messages: Observable<any[]>;
  message_target:string = "admin";
  constructor(public data:DataService, public game:GameService) { }

  ngOnInit(): void {
    this.messages = this.data.getMessages(this.game.game.id);
    this.messages.subscribe()
  }

}
