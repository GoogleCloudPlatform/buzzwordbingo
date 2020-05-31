import { Injectable,isDevMode } from '@angular/core';
import { AngularFirestore } from '@angular/fire/firestore'; 
import {GameService, Game} from '../service/game.service'

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
  constructor(  private firestore: AngularFirestore, 
                private game:GameService,
                ) { 


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


}