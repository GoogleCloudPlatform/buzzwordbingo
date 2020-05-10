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
  @Input() target:string;
  

  constructor() { }

  ngOnInit(): void {
  }

  ngAfterViewInit(): void {
    this.scrollDownWindow();
  }
  ngOnChanges(): void {
    this.scrollDownWindow();
  }

  findInAudience(message:Message){
    
    let self = this;
    let result:boolean = false;
    message.audience.forEach(function(val){
        if ((val == self.target) || (val == "all")) {
          result = true
        } 
      })

    return result;    
  }


  scrollDownWindow(){
    let d = document.querySelector('.messages');
    if (d) {
      d.scrollTop = d.scrollHeight;
    }
  }

}
