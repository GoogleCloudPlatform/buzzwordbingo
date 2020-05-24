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

  constructor(public auth:AuthService, public router:Router, route: ActivatedRoute,) {
    this.isAdmin = auth.isAdmin()
    this.id = route.snapshot.paramMap.get('id');
   }

  ngOnInit(): void {
  }

  logout(){
    this.auth.logout("user chose to logout")
  }

}
