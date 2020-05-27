import { Component, OnInit, Input } from '@angular/core';
import { GameService, Game } from 'src/app/service/game.service';
import { BehaviorSubject, Observable, of as observableOf  } from 'rxjs';
import { Router, ActivatedRoute } from '@angular/router';



@Component({
  selector: 'app-manage-admins',
  templateUrl: './manage-admins.component.html',
  styleUrls: ['./manage-admins.component.scss']
})
export class ManageAdminsComponent implements OnInit {
  public id:string;
  public game:BehaviorSubject<Game> = new BehaviorSubject(new Game);
  public gameid:string;
  constructor(private gameService:GameService, public router:Router, route: ActivatedRoute,) { 
    this.id = route.snapshot.paramMap.get('id');
    this.refreshGame();
  }

  refreshGame(){
    this.gameService.getGame(this.id).subscribe(val=>{let g:Game = val as Game; this.gameid=g.id; this.game.next(g);});
  }

  ngOnInit(): void {
  }

  onAdminAdd(email:string){
    this.gameService.addGameAdmin(this.gameid, email).subscribe();
    this.refreshGame();
  }

  onAdminRemove($event){
    console.log($event)
    $event.target.parentElement.style.display = 'none';
    this.gameService.removeGameAdmin(this.gameid, $event.target.id).subscribe(val=>{$event.target.parentElement.style.display = 'none';});
    this.refreshGame();
    
  }

}
