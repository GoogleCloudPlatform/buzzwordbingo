/// <reference path="../../../node_modules/@types/gapi/index.d.ts" />
/// <reference path="../../../node_modules/@types/gapi.auth2/index.d.ts" />
import { Injectable, isDevMode } from '@angular/core';
import { AngularFirestore } from '@angular/fire/firestore'; 
import {GameService, Game} from '../service/game.service'
import {GoogleAuthService} from '../service/googleauth.service'
import { AngularFireAuth } from '@angular/fire/auth';
import { Router } from '@angular/router';
import firebase from 'firebase/app';
import 'firebase/auth'; 
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

  constructor(private googleAuth:GoogleAuthService, private firestore: AngularFirestore, private game:GameService) { 
    googleAuth.login();
  }

  getPhrases() { 
    return this.firestore.collection("phrases").valueChanges();
  }

  getAdmins() { 
    return this.firestore.collection("admins").valueChanges();
  }

  getMessages(id:string, email:string) { 
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





}
