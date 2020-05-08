import { Component, OnInit, Input, Output, EventEmitter } from '@angular/core';
import { Observable, of } from 'rxjs';
import { Phrase} from '../../service/data.service'
import {GameService, Board} from '../../service/game.service'

@Component({
  selector: 'app-item',
  templateUrl: './item.component.html',
  styleUrls: ['./item.component.scss']
})

export class ItemComponent implements OnInit {
  @Input() phrase: Phrase;
  @Input() boardid: string;
  @Input() currentState:any;
  @Input() position: number;
  @Input() bingo:boolean = false;
  @Output() phraseEmitter = new EventEmitter<Phrase>();
  @Output() readyEmitter = new EventEmitter<ItemComponent>();
  disabled:boolean=false;
  
  constructor(private game:GameService) { }

  ngOnInit(): void {
    this.phrase.tid = this.convertPositionToTID(this.position);
  }

  ngAfterViewChecked(): void {
    if (this.phrase.selected){
      this.setDisplayChecked();
    }
  }

  ngAfterViewInit():void{
    this.readyEmitter.emit(this);
  } 

  select(){
    if (this.bingo){
      this.disabled = true;
      this.disable();
    }

    if (this.disabled){
      return;
    }
    
    this.selectDisplay();
    this.phraseEmitter.emit(this.phrase);
    this.game.record(this.phrase.id, this.boardid);
  }


  setDisplayChecked(){
    let item:HTMLElement = document.querySelector("#id_"+ this.phrase.id);
    let selectedPhraseCount = Object.keys(this.currentState).length;
    switch(selectedPhraseCount) {
      case 0:
        item.style.backgroundColor = "lavender";
        break;
      case 1:
        item.style.backgroundColor = "goldenrod";
        break;
      default:
        item.style.backgroundColor = "chartreuse";
    }
  }
  selectDisplay(){
    let item:HTMLElement = document.querySelector("#id_"+ this.phrase.id);

    if (this.phrase.selected){
      this.phrase.selected = false;
      item.style.backgroundColor = "";
    } else {
      this.phrase.selected = true;
      this.setDisplayChecked();
    }
    
    return;
  }

  public disable(){
    console.log("disabled called");
    this.disabled = true;
    let item:HTMLElement = document.querySelector("#id_"+ this.phrase.id);
    if (!this.phrase.selected){
      item.style.backgroundColor = "#DDD";
    }
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
