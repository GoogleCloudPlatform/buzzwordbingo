import { Component, OnInit } from '@angular/core';
import {Router} from '@angular/router';
import {AuthService, Player} from '../../service/auth.service'


@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent implements OnInit {

  constructor(private auth:AuthService, private router: Router) { }

  ngOnInit(): void {
  }

  submitPlayer(){
    console.log("called")
    let nameInput:HTMLInputElement = document.querySelector("#name") as HTMLInputElement;
    let emailInput:HTMLInputElement = document.querySelector("#email") as HTMLInputElement;
    this.auth.setPlayer(nameInput.value, emailInput.value);
    this.router.navigateByUrl('/game');

  }

}
