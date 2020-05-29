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
