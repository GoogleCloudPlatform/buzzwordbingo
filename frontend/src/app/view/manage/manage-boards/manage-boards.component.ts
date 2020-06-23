import { Component, OnInit, Input, ViewChild } from '@angular/core';
import { Observable, of } from 'rxjs';
import { DataService, Phrase} from '../../../service/data.service'
import {GameService, Board, Message, Record} from '../../../service/game.service'
import {ProgressbarComponent} from 'src/app/view/widgets/progressbar/progressbar.component'
import { FormalertComponent } from '../../widgets/formalert/formalert.component';

@Component({
  selector: 'app-manage-boards',
  templateUrl: './manage-boards.component.html',
  styleUrls: ['./manage-boards.component.scss']
})
export class ManageBoardsComponent implements OnInit {
  @ViewChild(ProgressbarComponent ) bar: ProgressbarComponent ; 
  @ViewChild(FormalertComponent ) formalert: FormalertComponent ;
  @Input() gid:string;
  public boards: Observable<any[]>;
  constructor(public data:DataService, public gameService:GameService) { }

  ngOnInit(): void {
    this.boards = this.data.getBoards(this.gid)
    this.boards.subscribe(val=>{this.bar.toggle()});
  }

  reset(bid:string, gid:string){
    this.gameService.resetboard(bid, gid);
    this.formalert.alert(`Board reset`);
  }

  onAdminAdd(email:string){
    this.gameService.addGameAdmin(this.gid, email).subscribe();
  }


}
