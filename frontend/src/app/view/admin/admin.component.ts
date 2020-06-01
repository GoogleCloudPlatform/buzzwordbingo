import { Component, OnInit } from '@angular/core';
import {GameService, Board, Message} from '../../service/game.service';
import { DataService, Phrase} from '../../service/data.service'
import { Observable, of } from 'rxjs';
import { Router, ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-admin',
  templateUrl: './admin.component.html',
  styleUrls: ['./admin.component.scss']
})
export class AdminComponent implements OnInit {
  public id:string;
  public messages: Observable<any[]>;
  message_target:string = "admin";
  constructor(public data:DataService, public game:GameService, public router:Router, route: ActivatedRoute,) {
    this.id = route.snapshot.paramMap.get('id');
   }

  ngOnInit(): void {
  }

}
