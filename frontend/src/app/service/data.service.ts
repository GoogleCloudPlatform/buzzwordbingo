import { Injectable,isDevMode } from '@angular/core';
import { AngularFirestore } from '@angular/fire/firestore'; 
import {GameService, Game} from '../service/game.service'
import { AngularFireAuth } from '@angular/fire/auth';
import * as firebase from 'firebase/app'
/// <reference types="gapi" />

export class Phrase{
  id:string
  text:string
  selected:boolean
  tid:string
  displayorder:number
}



@Injectable({
  providedIn: 'root'
})

export class DataService {
  GoogleAuth?: gapi.auth2.GoogleAuth = null;
  constructor(  private firestore: AngularFirestore, 
                private game:GameService,
                public fbauth: AngularFireAuth) { 

    this.initFirestoreAuth();

  }

  getPhrases() { 
    return this.firestore.collection("phrases").valueChanges();
  }

  getAdmins() { 
    return this.firestore.collection("admins").valueChanges();
  }

  getMessages(id) { 
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


  async initFirestoreAuth() {
    // The build is restricted by Cloud IAP on non-local environments. Google
    // API Client is used to take the id token from IAP's authentication and
    // auto authenticate Firebase.
    //
    // GAPI auth: https://developers.google.com/identity/sign-in/web/reference#gapiauth2authorizeparams-callback
    // GoogleAuthProvider: https://firebase.google.com/docs/reference/js/firebase.auth.GoogleAuthProvider

    if (isDevMode()) return;

    await this.loadGapiAuth();

    this.GoogleAuth = gapi.auth2.getAuthInstance();

    // Prevents a reauthentication and a redirect from `/signout` to `/dashboard` route
    if (this.GoogleAuth) {
      const token = this.GoogleAuth.currentUser.get().getAuthResponse().id_token;
      const credential = firebase.auth.GoogleAuthProvider.credential(token);
      this.fbauth.signInAndRetrieveDataWithCredential(credential)
    }

    this.login();
  }

  // Sign in button, which calls this method, should only be displayed for local
  // environment where Cloud IAP isn't setup
  login() {
    this.fbauth.useDeviceLanguage();
    const provider = new firebase.auth.GoogleAuthProvider();
    provider.addScope("profile");
    provider.addScope("email");
    this.fbauth.signInWithRedirect(provider);
  }


  private async loadGapiAuth() {
    let GAPI_CONFIG ={
      apiKey: '<API_KEY goes here>',
      clientId: "<CLIENT_ID goes here>",
      discoveryDocs: ["https://www.googleapis.com/discovery/v1/apis/drive/v3/rest"],
      scope: [
        "https://www.googleapis.com/auth/drive.metadata.readonly"
      ].join(" "),
    };
    await new Promise((resolve) => gapi.load('client:auth2', resolve));
    await new Promise((resolve) => gapi.auth2.init(GAPI_CONFIG).then(resolve));
  }

}
