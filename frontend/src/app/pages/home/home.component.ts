import { Component, inject, OnInit } from '@angular/core';
import { UserStateService } from '../../services/user-state/user-state.service';
import { AsyncPipe, JsonPipe } from '@angular/common';

@Component({
  selector: 'app-home',
  imports: [AsyncPipe, JsonPipe],
  templateUrl: './home.component.html',
  styleUrl: './home.component.scss',
})
export class HomeComponent implements OnInit {
  readonly userState = inject(UserStateService);
  readonly user$ = this.userState.user$;

  ngOnInit(): void {
    this.userState.user$.subscribe((user) => {
      if (user) {
        console.log('User data:', user);
      } else {
        console.log('No user data available');
      }
    });
  }
}
