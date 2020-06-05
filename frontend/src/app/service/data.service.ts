/// <reference path="../../../node_modules/@types/gapi/index.d.ts" />
/// <reference path="../../../node_modules/@types/gapi.auth2/index.d.ts" />
import { Injectable, isDevMode } from '@angular/core';
import { AngularFirestore } from '@angular/fire/firestore'; 
import {GameService, Game} from '../service/game.service'
import { AngularFireAuth } from '@angular/fire/auth';
import { Router } from '@angular/router';
import * as firebase from 'firebase'
import { environment } from '../../environments/environment';

declare var gapi:any;

export class Phrase{
  id:string
  text:string
  selected:boolean
  tid:string
  displayorder:number
}

const GAPI_CONFIG = {
  clientId: environment.client_id,
  fetch_basic_profile: true
}


@Injectable({
  providedIn: 'root'
})

export class DataService {

  constructor(public auth: AngularFireAuth, private firestore: AngularFirestore, private game:GameService) { 
    this.passCredentials();

  }

  getPhrases() { 
    return this.firestore.collection("phrases").valueChanges();
  }

  getAdmins() { 
    return this.firestore.collection("admins").valueChanges();
  }

  getMessages(id:string, email:string) { 
    console.log("email:", email)
    return this.firestore.collection("games").doc(id)
            .collection("messages", ref=> ref.where("audience", 'array-contains-any',
            ['all', email])).valueChanges();
  }

  getMessagesAdmin(id) { 
    return this.firestore.collection("games").doc(id).collection("messages").valueChanges();
  }

  getRecords(id) { 
    return this.firestore.collection("games").doc(id).collection("records").valueChanges();
  }

  getBoards(id) { 
    return this.firestore.collection("games").doc(id).collection("boards").valueChanges();
  }

  getGameBoard(gid:string, bid:string) { 
    return this.firestore.collection("games").doc(gid).collection("boards").doc(bid).collection("phrases").valueChanges();
  }



   passCredentials() {

    let self = this;

    gapi.load('client:auth2', function(){
      gapi.auth2.init(GAPI_CONFIG).then(function(googleAuth){

        if ( googleAuth.isSignedIn.get()){
          let token = googleAuth.currentUser.get().getAuthResponse().id_token;
          const credential = firebase.auth.GoogleAuthProvider.credential(token);
          self.auth.signInWithCredential(credential);
        } else {
          googleAuth.signIn().then((guser) =>{
            const token = guser.getAuthResponse().id_token;
            const credential = firebase.auth.GoogleAuthProvider.credential(token);
            self.auth.signInWithCredential(credential);
          })
        }
      }, function(err){console.log(err);})

    })

  }


}
