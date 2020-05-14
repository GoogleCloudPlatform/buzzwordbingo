import { Component, OnInit, Input } from '@angular/core';
import { DataService, Phrase} from '../../service/data.service'

@Component({
  selector: 'app-admin-phrase-form',
  templateUrl: './admin-phrase-form.component.html',
  styleUrls: ['./admin-phrase-form.component.scss']
})
export class AdminPhraseFormComponent implements OnInit {
  @Input() phrase: Phrase;
  constructor(public data: DataService) { }

  ngOnInit(): void {
  }

  onPhraseSubmit(){
    this.data.updatePhrase(this.phrase);
  }

}
