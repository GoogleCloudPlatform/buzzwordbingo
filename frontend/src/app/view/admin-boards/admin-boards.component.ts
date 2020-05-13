import { Component, OnInit } from '@angular/core';
import { Observable, of } from 'rxjs';
import { DataService, Phrase} from '../../service/data.service'
import {GameService, Board, Message, Record} from '../../service/game.service'

@Component({
  selector: 'app-admin-boards',
  templateUrl: './admin-boards.component.html',
  styleUrls: ['./admin-boards.component.scss']
})
export class AdminBoardsComponent implements OnInit {

  public boards: Observable<any[]>;
  constructor(public data:DataService, public game:GameService) { }

  ngOnInit(): void {
    this.boards = this.data.getBoards(this.game.game.id)
    this.boards.subscribe(ref=>console.log(ref));
  }

  reset(bid:string){
    console.log("Reset called")
    this.game.resetboard(bid);
  }

}
