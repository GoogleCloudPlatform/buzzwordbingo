import { Injectable } from '@angular/core';
import { AngularFirestore } from '@angular/fire/firestore'; 
import {GameService, Game} from '../service/game.service'

export class Phrase{
  id:string
  text:string
  selected:boolean
  tid:string
}


@Injectable({
  providedIn: 'root'
})

export class DataService {

  constructor(private firestore: AngularFirestore, private game:GameService) { }

  getPhrases() { 
    return this.firestore.collection("phrases").valueChanges();
  }

  getMessages(id) { 
    return this.firestore.collection("games").doc(id).collection("messages").valueChanges();
  }

  getRecords(id) { 
    return this.firestore.collection("games").doc(id).collection("records").valueChanges();
  }

}
