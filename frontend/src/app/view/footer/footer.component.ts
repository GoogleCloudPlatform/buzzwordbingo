import { Component, OnInit } from '@angular/core';
import { ThemeService } from 'src/app/service/theme.service';

@Component({
  selector: 'app-footer',
  templateUrl: './footer.component.html',
  styleUrls: ['./footer.component.scss']
})
export class FooterComponent implements OnInit {
  public theme:string="light"
  constructor(private themeService:ThemeService) { 
    let theme = localStorage.getItem('theme');
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
        localStorage.setItem("theme", "unicorn");
        this.theme ="unicorn";
        break;
      case "dark":
        this.themeService.toggleDark()
        localStorage.setItem("theme", "dark");
        this.theme ="dark";
        break;
      default:
        this.themeService.toggleLight()
        localStorage.setItem("theme", "light");
        this.theme ="light";
    }
  }

}
