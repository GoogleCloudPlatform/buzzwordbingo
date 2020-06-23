import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-formalert',
  templateUrl: './formalert.component.html',
  styleUrls: ['./formalert.component.scss']
})
export class FormalertComponent implements OnInit {
  public message: string ="Just a little thing to show";
  public display:boolean = false;
  public cssclass:string = "hide";
  constructor() { }

  ngOnInit(): void {
  }

  alert(message:string){
      let self = this;
      this.message = message;
      this.cssclass = "info";
      setTimeout(function () {self.close()}, 4000);
  }

  error(message:string){
    let self = this;
    this.message = message;
    this.cssclass = "error";
    setTimeout(function () {self.close()}, 4000);
}

  close(){
      this.cssclass = "hide";
  }

}
