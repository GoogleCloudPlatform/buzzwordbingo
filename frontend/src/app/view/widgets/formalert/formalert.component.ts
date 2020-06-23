/**
 * Copyright 2020 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
