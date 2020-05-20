import { Component, OnInit, Input } from '@angular/core';
import { DataService, Phrase} from '../../service/data.service'

@Component({
  selector: 'app-admin-phrase-form',
  templateUrl: './admin-phrase-form.component.html',
  styleUrls: ['./admin-phrase-form.component.scss']
})
export class AdminPhraseFormComponent implements OnInit {
  @Input() phrase: Phrase;
  timeout = null;
  constructor(public data: DataService) { }

  ngOnInit(): void {
  }

  keyup(){
    let self = this;
    clearTimeout(this.timeout);

    this.timeout = setTimeout(function () {self.onPhraseSubmit()}, 1000);
  }

  onPhraseSubmit(){
    this.data.updatePhrase(this.phrase);
  }

}
