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
