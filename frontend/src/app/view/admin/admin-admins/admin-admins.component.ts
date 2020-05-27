import { Component, OnInit, Input } from '@angular/core';
import { GameService, Game } from 'src/app/service/game.service';
import { BehaviorSubject, Observable, of as observableOf  } from 'rxjs';
import { Player} from 'src/app/service/auth.service'

@Component({
  selector: 'app-admin-admins',
  templateUrl: './admin-admins.component.html',
  styleUrls: ['./admin-admins.component.scss']
})
export class AdminAdminsComponent implements OnInit {
  public admins:BehaviorSubject<Player[]> = new BehaviorSubject([]);
  constructor(private gameService:GameService) { 
    this.refreshAdmins();
  }
  ngOnInit(): void {
  }

  refreshAdmins(){
    this.gameService.getAdmins().subscribe(val=>{let p:Player[] = val as Player[]; console.log(val); this.admins.next(p);});
  }

  onAdminAdd(email:string){
    this.gameService.addAdmin(email).subscribe(val=>{this.refreshAdmins()});
    
  }

  onAdminRemove($event){
    console.log($event)
    $event.target.parentElement.style.display = 'none';
    this.gameService.removeAdmin($event.target.id).subscribe(val=>{$event.target.parentElement.style.display = 'none'; this.refreshAdmins();});
    
    
  }

}
