import { Component, OnInit, Input } from '@angular/core';
import {GameService, Record, Game} from '../../../../service/game.service';
@Component({
  selector: 'app-manage-phrase-form',
  templateUrl: './manage-phrase-form.component.html',
  styleUrls: ['./manage-phrase-form.component.scss']
})
export class ManagePhraseFormComponent implements OnInit {
  @Input() record:Record;
  @Input() game:Game;
  timeout = null;
  constructor(private gameService:GameService) { }

  ngOnInit(): void {
  }

  keyup(){
    let self = this;
    clearTimeout(this.timeout);
    this.timeout = setTimeout(function () {self.onPhraseSubmit()}, 1000);
  }

  onPhraseSubmit(){
    clearTimeout(this.timeout);
    console.log("Save Phrase")
    this.gameService.updateGamePhrase(this.game.id, this.record.phrase).subscribe();
  }

}
