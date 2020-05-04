import { Component, OnInit, Input, Output, EventEmitter } from '@angular/core';
import { Observable, of } from 'rxjs';
import { Phrase} from '../../service/data.service'

@Component({
  selector: 'app-item',
  templateUrl: './item.component.html',
  styleUrls: ['./item.component.scss']
})

export class ItemComponent implements OnInit {
  @Input() phrase: Phrase;
  @Input() position: number;
  @Output() phraseEmitter = new EventEmitter<Phrase>();
  
  constructor() { }

  ngOnInit(): void {
    this.phrase.tid = this.convertPositionToTID(this.position);
  }

  select(){

    let id = this.phrase.id;


    if (this.phrase.selected){
      document.querySelector("#id_"+ id).classList.remove("selected");
      this.phrase.selected = false;
    } else {
      document.querySelector("#id_"+ id).classList.add("selected");
      this.phrase.selected = true;
    }
    

    this.phraseEmitter.emit(this.phrase);
    return;
  }

  

  convertPositionToTID(position){
    let first = "";
    let second = "";
    second = Math.ceil(position/5).toString();
    switch(position % 5) {
      case 1:
        first = "B";
        break;
      case 2:
        first = "I";
        break;
      case 3:
        first = "N";
        break;
      case 4:
        first = "G";
        break;  
      default:
        first = "O";
    }
    return first + second;

  }

}
