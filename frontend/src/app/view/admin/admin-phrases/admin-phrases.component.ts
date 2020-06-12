import { Component, OnInit, ViewChild } from '@angular/core';
import { Observable, of } from 'rxjs';
import { DataService, Phrase} from '../../../service/data.service'
import {GameService, Board, Message, Record} from '../../../service/game.service'
import { ProgressspinnerComponent } from '../../widgets/progressspinner/progressspinner.component';



@Component({
  selector: 'app-admin-phrases',
  templateUrl: './admin-phrases.component.html',
  styleUrls: ['./admin-phrases.component.scss']
})
export class AdminPhrasesComponent implements OnInit {
  @ViewChild(ProgressspinnerComponent ) spinner: ProgressspinnerComponent ;
  public phrases: Observable<any[]>;
  
  constructor(public data:DataService, public game:GameService) { }

  ngOnInit(): void {

    this.phrases = this.data.getPhrases()
    this.phrases.subscribe(val=>{this.spinner.toggle()});
  }

  

}
