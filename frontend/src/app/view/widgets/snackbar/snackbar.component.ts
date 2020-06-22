import { Component, OnInit, Input } from '@angular/core';

@Component({
  selector: 'app-snackbar',
  templateUrl: './snackbar.component.html',
  styleUrls: ['./snackbar.component.scss']
})
export class SnackbarComponent implements OnInit {
  public message: string;
  public display:boolean = false;
  constructor() { }

  ngOnInit(): void {
    this.message = ""
  }

  show(message:string){
    let self = this;
    this.message = message;
    this.display = true;
    setTimeout(function () {self.close()}, 8000);
  }

  close()  {
    this.display = false;
  }

}
