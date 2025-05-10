// src/app/services/session/session.service.ts
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import * as CryptoJS from 'crypto-js';
import { environment } from '../../../environments/environment';

interface JwtPayload {
  exp?: number;
  iat?: number;
  [key: string]: any;
}

@Injectable({ providedIn: 'root' })
export class SessionService {
  private readonly COOKIE_NAME = 'auth_token';
  private readonly authState = new BehaviorSubject<boolean>(
    this.hasValidToken()
  );
  readonly isAuthenticated$: Observable<boolean> =
    this.authState.asObservable();

  private expirationTimerId?: ReturnType<typeof setTimeout>;

  private readonly secretKey = CryptoJS.enc.Utf8.parse(environment.tokenSecret);
  private readonly iv = CryptoJS.enc.Utf8.parse(environment.tokenIv);

  constructor() {
    // Au démarrage, planifier expiration si déjà connecté
    if (this.hasValidToken()) {
      this.scheduleExpiration();
    }
  }

  setToken(token: string): void {
    const encrypted = CryptoJS.AES.encrypt(token, this.secretKey, {
      iv: this.iv,
    }).toString();
    const payload = this.decodeToken<JwtPayload>(token);
    if (!payload?.exp) {
      throw new Error('Token JWT sans claim exp');
    }
    const expiresAt = new Date(payload.exp * 1000);
    this.setCookie(this.COOKIE_NAME, encrypted, expiresAt);
    this.authState.next(true);
    this.scheduleExpiration();
  }

  getToken(): string | null {
    const cipher = this.getCookie(this.COOKIE_NAME);
    if (!cipher) return null;
    try {
      const bytes = CryptoJS.AES.decrypt(cipher, this.secretKey, {
        iv: this.iv,
      });
      const plain = bytes.toString(CryptoJS.enc.Utf8);
      return plain || null;
    } catch {
      return null;
    }
  }

  clearSession(): void {
    this.deleteCookie(this.COOKIE_NAME);
    this.authState.next(false);
    this.clearExpirationTimer();
  }

  getUserId(): string | null {
    const payload = this.getPayload();
    return payload?.['user_id'] ?? null;
  }

  getPayload(): JwtPayload | null {
    const token = this.getToken();
    return token ? this.decodeToken<JwtPayload>(token) : null;
  }

  getExpirationDate(): Date | null {
    const payload = this.getPayload();
    return payload?.exp ? new Date(payload.exp * 1000) : null;
  }

  isTokenExpired(): boolean {
    return !this.hasValidToken();
  }

  private hasValidToken(): boolean {
    const payload = this.getPayload();
    return !!(payload?.exp && Date.now() < payload.exp * 1000);
  }

  private scheduleExpiration(): void {
    this.clearExpirationTimer();
    const payload = this.getPayload();
    if (!payload?.exp) return;

    const ms = payload.exp * 1000 - Date.now();
    if (ms <= 0) {
      this.clearSession();
    } else {
      this.expirationTimerId = setTimeout(() => this.clearSession(), ms);
    }
  }

  private clearExpirationTimer(): void {
    if (this.expirationTimerId) {
      clearTimeout(this.expirationTimerId);
      this.expirationTimerId = undefined;
    }
  }

  private decodeToken<T = any>(token: string): T | null {
    try {
      const [, b64] = token.split('.');
      const bin = atob(b64.replace(/-/g, '+').replace(/_/g, '/'));
      return JSON.parse(bin) as T;
    } catch {
      return null;
    }
  }

  private setCookie(name: string, value: string, expires: Date): void {
    const cookie = [
      `${name}=${encodeURIComponent(value)}`,
      `expires=${expires.toUTCString()}`,
      `path=/`,
      `Secure`,
      `SameSite=Strict`,
    ].join('; ');
    document.cookie = cookie;
  }

  private getCookie(name: string): string | null {
    const match = document.cookie
      .split('; ')
      .find((row) => row.startsWith(name + '='));
    return match ? decodeURIComponent(match.split('=')[1]) : null;
  }

  private deleteCookie(name: string): void {
    document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/`;
  }
}
