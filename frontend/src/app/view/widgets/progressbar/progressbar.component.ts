import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-progressbar',
  templateUrl: './progressbar.component.html',
  styleUrls: ['./progressbar.component.scss']
})
export class ProgressbarComponent implements OnInit {
  public display:boolean = true;
  constructor() { }

  ngOnInit(): void {
  }

  public toggle(){
    this.display = false;
  }

}
