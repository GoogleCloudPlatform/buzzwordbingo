import { Component, OnInit, Input } from '@angular/core';
import { Observable, of } from 'rxjs';
import { DataService, Phrase} from '../../../service/data.service'
import {GameService, Board, Message, Record} from '../../../service/game.service'

@Component({
  selector: 'app-admin-master',
  templateUrl: './admin-master.component.html',
  styleUrls: ['./admin-master.component.scss']
})
export class AdminMasterComponent implements OnInit {
  @Input() id:string;
  public records: Observable<any[]>;
  constructor(public data:DataService, public game:GameService) { }

  ngOnInit(): void {
     this.records = this.data.getRecords(this.id)
     this.records.subscribe();
  }

  playerCount(record:Record):number{
    if (record.players != null){
      return record.players.length
    }
    return 0;
  }

  isActive(record:Record):boolean{
    if (record.players != null && record.players.length > 0){
      return true
    }
    return false;
  }

}
