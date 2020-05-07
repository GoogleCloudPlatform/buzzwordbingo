import { Injectable } from '@angular/core';
import { AngularFirestore } from '@angular/fire/firestore';

export class Phrase{
  id:number
  text:string
  selected:boolean
  tid:string
}


@Injectable({
  providedIn: 'root'
})

export class DataService {

  constructor(private firestore: AngularFirestore) { }

  getPhrases() { 
    return this.firestore.collection("phrases").valueChanges();
  }

}
