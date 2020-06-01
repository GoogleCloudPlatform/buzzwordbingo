import { Component } from '@angular/core';
import { Meta } from '@angular/platform-browser';
import { environment } from '../environments/environment';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  title = 'bingomeeting';

  constructor(public meta: Meta) { 
    this.meta.updateTag({ name: 'google-signin-client_id', content: environment.client_id });
  }
}
