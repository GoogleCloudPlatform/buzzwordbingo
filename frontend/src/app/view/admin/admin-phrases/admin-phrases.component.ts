import { Component, OnInit } from '@angular/core';
import { Observable, of } from 'rxjs';
import { DataService, Phrase} from '../../../service/data.service'
import {GameService, Board, Message, Record} from '../../../service/game.service'

@Component({
  selector: 'app-admin-phrases',
  templateUrl: './admin-phrases.component.html',
  styleUrls: ['./admin-phrases.component.scss']
})
export class AdminPhrasesComponent implements OnInit {
  public phrases: Observable<any[]>;
  
  constructor(public data:DataService, public game:GameService) { }

  ngOnInit(): void {

    this.phrases = this.data.getPhrases()
    this.phrases.subscribe();
  }

  

}
