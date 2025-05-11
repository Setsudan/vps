import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { SessionService } from '../session/session.service';
import { map, Observable, switchMap, tap } from 'rxjs';
import { environment } from '../../../environments/environment';
import { APIResponse, unwrapAPIResponse } from '../../../types/apiResponse';
import { UserStateService } from '../user-state/user-state.service';
import { IUser } from '../../../types/user';

interface RegisterPayload {
  username: string;
  email: string;
  password: string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly session = inject(SessionService);
  private readonly userState = inject(UserStateService);
  private readonly apiUrl = environment.apiUrl;

  /**
   * Appelle POST /auth/register
   * @returns Observable<void> (201 Created ou erreur)
   */
  register(payload: RegisterPayload): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/auth/register`, payload);
  }

  /**
   * Appelle POST /auth/login, stocke le token JWT en session
   * @returns Observable<void> (200 OK ou erreur)
   */
  login(email: string, password: string): Observable<void> {
    return this.http
      .post<APIResponse<{ token: string }>>(`${this.apiUrl}/auth/login`, {
        email,
        password,
      })
      .pipe(
        map(unwrapAPIResponse),
        tap((data) => this.session.setToken(data.token)),
        switchMap(() =>
          this.http.get<APIResponse<IUser>>(`${this.apiUrl}/user/me`)
        ),
        map(unwrapAPIResponse),
        tap((user) => this.userState.set(user)),
        map(() => void 0)
      );
  }

  /**
   * Logout : efface la session côté client
   */
  logout(): void {
    this.session.clearSession();
    this.userState.clear();
  }

  /**
   * Observable pour savoir si l'utilisateur est connecté
   */
  get isAuthenticated$(): Observable<boolean> {
    return this.session.isAuthenticated$;
  }
}
