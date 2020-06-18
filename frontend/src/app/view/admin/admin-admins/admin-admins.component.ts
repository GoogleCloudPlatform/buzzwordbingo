import { Component, OnInit, Input,ViewChild } from '@angular/core';
import { GameService, Game } from 'src/app/service/game.service';
import { BehaviorSubject, Observable, of as observableOf  } from 'rxjs';
import { Player} from 'src/app/service/auth.service'
import { ProgressspinnerComponent } from '../../widgets/progressspinner/progressspinner.component';


@Component({
  selector: 'app-admin-admins',
  templateUrl: './admin-admins.component.html',
  styleUrls: ['./admin-admins.component.scss']
})
export class AdminAdminsComponent implements OnInit {
  @ViewChild(ProgressspinnerComponent ) spinner: ProgressspinnerComponent ;
  public admins:BehaviorSubject<Player[]> = new BehaviorSubject([]);
  constructor(private gameService:GameService) { 
    this.refreshAdmins();
  }
  ngOnInit(): void {
  }

  refreshAdmins(){
    this.gameService.getAdmins().subscribe(val=>{
        let p:Player[] = val as Player[]; 
        this.admins.next(p);
        this.spinner.toggle();
    });
  }

  onAdminAdd(email:string){
    this.gameService.addAdmin(email).subscribe(val=>{this.refreshAdmins()});
  }

  onAdminRemove($event, email:string ){
    $event.target.parentElement.style.display = 'none';
    this.gameService.removeAdmin(email).subscribe(val=>{$event.target.parentElement.style.display = 'none'; this.refreshAdmins();});
  }

}
