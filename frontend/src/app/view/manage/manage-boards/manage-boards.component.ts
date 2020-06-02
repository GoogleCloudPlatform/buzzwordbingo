import { Component, OnInit, Input } from '@angular/core';
import { Observable, of } from 'rxjs';
import { DataService, Phrase} from '../../../service/data.service'
import {GameService, Board, Message, Record} from '../../../service/game.service'

@Component({
  selector: 'app-manage-boards',
  templateUrl: './manage-boards.component.html',
  styleUrls: ['./manage-boards.component.scss']
})
export class ManageBoardsComponent implements OnInit {
  @Input() gid:string;
  public boards: Observable<any[]>;
  constructor(public data:DataService, public gameService:GameService) { }

  ngOnInit(): void {
    this.boards = this.data.getBoards(this.gid)
    this.boards.subscribe();
  }

  reset(bid:string, gid:string){
    this.gameService.resetboard(bid, gid);
  }

  onAdminAdd(email:string){
    this.gameService.addGameAdmin(this.gid, email).subscribe();
  }


}
