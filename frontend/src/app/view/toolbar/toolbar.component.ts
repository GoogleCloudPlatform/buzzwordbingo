import { Component, OnInit } from '@angular/core';
import { AuthService } from 'src/app/service/auth.service';
import {Router} from '@angular/router';

@Component({
  selector: 'app-toolbar',
  templateUrl: './toolbar.component.html',
  styleUrls: ['./toolbar.component.scss']
})
export class ToolbarComponent implements OnInit {

  isAdmin:boolean = false;

  constructor(public auth:AuthService, public router:Router) {
    this.isAdmin = auth.isAdmin()
    console.log("Isadmin");
   }

  ngOnInit(): void {
  }

  logout(){
    localStorage.clear();
    this.router.navigateByUrl('/login');
  }

}
