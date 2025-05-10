import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { SessionService } from '../session/session.service';
import { map, Observable, tap } from 'rxjs';
import { environment } from '../../../environments/environment';
import { APIResponse, unwrapAPIResponse } from '../../../types/apiResponse';

interface RegisterPayload {
  username: string;
  email: string;
  password: string;
}

interface LoginPayload {
  token: string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly session = inject(SessionService);
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
      .post<APIResponse<LoginPayload>>(`${this.apiUrl}/auth/login`, {
        email,
        password,
      })
      .pipe(
        map(unwrapAPIResponse),
        tap((data) => {
          this.session.setToken(data.token);
        }),
        map(() => void 0)
      );
  }

  /**
   * Logout : efface la session côté client
   */
  logout(): void {
    this.session.clearSession();
  }

  /**
   * Observable pour savoir si l'utilisateur est connecté
   */
  get isAuthenticated$(): Observable<boolean> {
    return this.session.isAuthenticated$;
  }
}
