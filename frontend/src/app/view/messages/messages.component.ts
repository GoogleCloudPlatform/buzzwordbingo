import { Component, OnInit, Input, ChangeDetectorRef, AfterContentChecked} from '@angular/core';
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
  

  constructor(private cdref: ChangeDetectorRef) { }

  ngOnInit(): void {
    this.scrollDownWindow();
    this.messages.subscribe(val=>{this.scrollDownWindow();})
  }

  ngAfterViewInit(): void {
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
    this.cdref.detectChanges();
  }


}
