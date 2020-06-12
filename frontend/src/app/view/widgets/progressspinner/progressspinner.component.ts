import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-progressspinner',
  templateUrl: './progressspinner.component.html',
  styleUrls: ['./progressspinner.component.scss']
})
export class ProgressspinnerComponent implements OnInit {
  public display:boolean = true;
  constructor() { }

  ngOnInit(): void {
  }

  public toggle(){
    this.display = false;
  }

}
