import { Component, OnInit } from '@angular/core';
import { AuthService } from 'src/app/service/auth.service';
import { Router, ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-toolbar',
  templateUrl: './toolbar.component.html',
  styleUrls: ['./toolbar.component.scss']
})
export class ToolbarComponent implements OnInit {

  isAdmin:boolean = false;
  public games:any;
  public id:string;
  public currentRoute:string;

  constructor(public auth:AuthService, public router:Router, route: ActivatedRoute,) {
    this.isAdmin = auth.isAdmin()
    this.id = route.snapshot.paramMap.get('id');
    this.currentRoute = this.router.url; 
    if (this.id != null ){
      this.auth.checkGameAdmin(this.id);
    }
    
   }

  ngOnInit(): void {
  }

  isLogin(){
    if (this.currentRoute.includes("invite")){
      return true;
    }
    if (this.currentRoute.includes("login")){
      return true;
    }
    
    return false;
  }

  logout(){
    this.auth.logout("user chose to logout")
  }

}
