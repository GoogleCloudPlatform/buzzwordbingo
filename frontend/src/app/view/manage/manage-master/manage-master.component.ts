import { Component, OnInit, Input, ViewChild } from '@angular/core';
import { Observable, of } from 'rxjs';
import { DataService, Phrase} from '../../../service/data.service'
import {GameService, Board, Message, Record} from '../../../service/game.service'
import {ProgressbarComponent} from 'src/app/view/widgets/progressbar/progressbar.component'


@Component({
  selector: 'app-manage-master',
  templateUrl: './manage-master.component.html',
  styleUrls: ['./manage-master.component.scss']
})
export class ManageMasterComponent implements OnInit {
  @ViewChild(ProgressbarComponent ) bar: ProgressbarComponent ; 
  @Input() id:string;
  public records: Observable<any[]>;
  constructor(public data:DataService, public game:GameService) { }

  ngOnInit(): void {
     this.records = this.data.getRecords(this.id)
     this.records.subscribe(val=>{this.bar.toggle()});
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
