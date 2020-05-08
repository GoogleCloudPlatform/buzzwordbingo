import { Component, OnInit, Input} from '@angular/core';
import {GameService, Message} from '../../service/game.service'
import { Observable } from 'rxjs';


@Component({
  selector: 'app-messages',
  templateUrl: './messages.component.html',
  styleUrls: ['./messages.component.scss']
})
export class MessagesComponent implements OnInit {
  @Input() messages: Observable<any>;

  constructor() { }

  ngOnInit(): void {
  }

}
