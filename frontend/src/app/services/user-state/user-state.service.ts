// src/app/services/user-state/user-state.service.ts
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import * as CryptoJS from 'crypto-js';
import { SessionService } from '../session/session.service';
import { environment } from '../../../environments/environment';
import { IUser } from '../../../types/user';

interface WrappedUser {
  user: IUser;
  expiresAt: string; // ISO
}

@Injectable({ providedIn: 'root' })
export class UserStateService {
  private readonly STORAGE_KEY = 'current_user_data';
  private readonly secretKey = CryptoJS.enc.Utf8.parse(environment.tokenSecret);
  private readonly iv = CryptoJS.enc.Utf8.parse(environment.tokenIv);

  private expirationTimer?: number;
  private readonly subject = new BehaviorSubject<IUser | null>(null);
  readonly user$: Observable<IUser | null> = this.subject.asObservable();

  constructor(private readonly session: SessionService) {
    this.loadFromStorage();
  }

  private loadFromStorage(): void {
    const cipher = localStorage.getItem(this.STORAGE_KEY);
    if (!cipher) {
      return;
    }
    try {
      const bytes = CryptoJS.AES.decrypt(cipher, this.secretKey, {
        iv: this.iv,
      });
      const text = bytes.toString(CryptoJS.enc.Utf8);
      const { user, expiresAt } = JSON.parse(text) as WrappedUser;
      const expDate = new Date(expiresAt);

      if (Date.now() >= expDate.getTime()) {
        this.clear();
        return;
      }

      this.subject.next(user);
      this.scheduleClear(expDate);
    } catch {
      this.clear();
    }
  }

  set(user: IUser): void {
    // choose expiration = token exp or a fixed TTL
    const tokenExp = this.session.getExpirationDate();
    const expiresAt =
      tokenExp ?? new Date(Date.now() + environment.userCacheTTL * 1000);

    const wrapper: WrappedUser = {
      user,
      expiresAt: expiresAt.toISOString(),
    };
    const plain = JSON.stringify(wrapper);
    const cipher = CryptoJS.AES.encrypt(plain, this.secretKey, {
      iv: this.iv,
    }).toString();

    localStorage.setItem(this.STORAGE_KEY, cipher);
    this.subject.next(user);

    this.clearTimer();
    this.scheduleClear(expiresAt);
  }

  get(): IUser | null {
    return this.subject.value;
  }

  clear(): void {
    localStorage.removeItem(this.STORAGE_KEY);
    this.subject.next(null);
    this.clearTimer();
  }

  private scheduleClear(at: Date) {
    const ms = at.getTime() - Date.now();
    if (ms <= 0) {
      this.clear();
      return;
    }
    this.expirationTimer = window.setTimeout(() => this.clear(), ms);
  }

  private clearTimer() {
    if (this.expirationTimer != null) {
      clearTimeout(this.expirationTimer);
      this.expirationTimer = undefined;
    }
  }
}
