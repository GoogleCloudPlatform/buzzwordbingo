import { Component, OnInit, Input } from '@angular/core';
import { DataService} from '../../../../service/data.service'
import { GameService, Phrase} from '../../../../service/game.service'

@Component({
  selector: 'app-admin-phrase-form',
  templateUrl: './admin-phrase-form.component.html',
  styleUrls: ['./admin-phrase-form.component.scss']
})
export class AdminPhraseFormComponent implements OnInit {
  @Input() phrase: Phrase;
  timeout = null;
  constructor(public gameService: GameService) { }

  ngOnInit(): void {
  }

  keyup(){
    let self = this;
    clearTimeout(this.timeout);
    this.timeout = setTimeout(function () {self.onPhraseSubmit()}, 1000);
  }

  onPhraseSubmit(){
    clearTimeout(this.timeout);
    this.gameService.updateMasterPhrase(this.phrase).subscribe();
  }

}
