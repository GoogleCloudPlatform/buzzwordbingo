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
import { ThemeService } from 'src/app/service/theme.service';
import { LocalstorageService } from 'src/app/service/localstorage.service';

@Component({
  selector: 'app-footer',
  templateUrl: './footer.component.html',
  styleUrls: ['./footer.component.scss']
})
export class FooterComponent implements OnInit {
  public theme:string="light"
  constructor(private themeService:ThemeService, private localStorageService:LocalstorageService) { 
    let theme = localStorageService.getTheme();
    if (theme != null){
      this.click(theme);
    }
  }

  ngOnInit(): void {
  }

  click(theme:string){
    switch(theme) {
      case "unicorn":
        this.themeService.toggleUnicorn()
        this.theme ="unicorn";
        break;
      case "dark":
        this.themeService.toggleDark()
        this.theme ="dark";
        break;
      default:
        this.themeService.toggleLight()
        this.theme ="light";
    }
    this.localStorageService.setTheme(this.theme);
  }

}
