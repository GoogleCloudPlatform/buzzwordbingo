import { Component, OnInit, Input, Output, EventEmitter, ViewChild } from '@angular/core';
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
  }

  ngAfterViewChecked(): void {
    if (this.phrase.selected){
      this.setDisplayChecked();
    }
  }

  ngAfterViewInit():void{
    this.readyEmitter.emit(this);
  } 

  ngOnChange(){
    console.log("Bingo:", this.bingo);
  }

  select(){
    if (this.bingo){
      this.disabled = true;
      this.disable();
    }

    if (this.phrase.text == "FREE"){
      return;
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
    item.classList.add("selected");
    switch(selectedPhraseCount) {
    case 0:
      item.style.backgroundColor = '#f4d9ff';
      item.style.color = '#3f3d40';
      break;
    case 1:
      item.style.backgroundColor = '#f3ffdd';
      item.style.color = '#3f3d40';
      break;
    case 2:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#3f3d40';
      break;
    case 3:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#ddffeb';
      break;
    case 4:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#bddbff';
      break;
    case 5:
      item.style.backgroundColor = '#f4d9ff';
      item.style.color = '#3f3d40';
      break;
    case 6:
      item.style.backgroundColor = '#f3ffdd';
      item.style.color = '#3f3d40';
      break;
    case 7:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#3f3d40';
      break;
    case 8:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#ddffeb';
      break;
    case 9:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#bddbff';
      break;
    case 10:
      item.style.backgroundColor = '#f4d9ff';
      item.style.color = '#3f3d40';
      break;
    case 11:
      item.style.backgroundColor = '#f3ffdd';
      item.style.color = '#3f3d40';
      break;
    case 12:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#3f3d40';
      break;
    case 13:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#ddffeb';
      break;
    case 14:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#bddbff';
      break;
    case 15:
      item.style.backgroundColor = '#f4d9ff';
      item.style.color = '#3f3d40';
      break;
    case 16:
      item.style.backgroundColor = '#f3ffdd';
      item.style.color = '#3f3d40';
      break;
    case 17:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#3f3d40';
      break;
    case 18:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#ddffeb';
      break;
    case 19:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#bddbff';
      break;
    case 20:
      item.style.backgroundColor = '#f4d9ff';
      item.style.color = '#3f3d40';
      break;
    case 21:
      item.style.backgroundColor = '#f3ffdd';
      item.style.color = '#3f3d40';
      break;
    case 22:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#3f3d40';
      break;
    case 23:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#ddffeb';
      break;
    case 24:
      item.style.backgroundColor = '#ffeedd';
      item.style.color = '#bddbff';
      break;

    default:
      item.style.backgroundColor = '#f3ffdd';
      break;

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
    this.disabled = true;
    console.log("Disable!");
    let item:HTMLElement = document.querySelector("#id_"+ this.phrase.id);
    if (!this.phrase.selected){
      item.classList.add("disabled");
    }
    item.classList.add("board-disabled");
  }

  


}
