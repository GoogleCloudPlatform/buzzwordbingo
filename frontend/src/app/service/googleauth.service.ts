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

import { Injectable } from '@angular/core';
import { environment } from '../../environments/environment';
import firebase from 'firebase/app';
import { AngularFireAuth } from '@angular/fire/auth';
import { AngularFirestore } from '@angular/fire/firestore'; 
import { Observable, of, BehaviorSubject  } from 'rxjs';

declare var gapi:any;

const GAPI_CONFIG = {
  clientId: environment.client_id,
  fetch_basic_profile: true
}

@Injectable({
  providedIn: 'root'
})
export class GoogleAuthService {

  private googleAuth:any;
  public authed:BehaviorSubject<boolean> = new BehaviorSubject<boolean>(null);

  constructor(public auth: AngularFireAuth, private firestore: AngularFirestore) { 
  }

  login() {
    let self = this;

    gapi.load('client:auth2', function(){
      gapi.auth2.init(GAPI_CONFIG).then(function(ga){
        self.googleAuth = ga;
        if ( self.googleAuth.isSignedIn.get()){
          let token = self.googleAuth.currentUser.get().getAuthResponse().id_token;
          const credential = firebase.auth.GoogleAuthProvider.credential(token);
          self.auth.signInWithCredential(credential);
          self.authed.next(true);
        } else {
          self.googleAuth.signIn().then((guser) =>{
            const token = guser.getAuthResponse().id_token;
            const credential = firebase.auth.GoogleAuthProvider.credential(token);
            self.auth.signInWithCredential(credential);
            self.authed.next(true);
          }, err=>{console.log("Firebase auth error",err)})
        }
      }, function(err){console.log("Google Auth error:", err);})

    })

  }

  logout() {
    this.googleAuth.signOut();
  }
}
