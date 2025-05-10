// src/app/services/session/session.service.spec.ts
import { SessionService } from './session.service';
import { environment } from '../../../environments/environment';
import * as CryptoJS from 'crypto-js';
import { take } from 'rxjs/operators';

describe('SessionService', () => {
  let service: SessionService;
  let cookieStore: Record<string, string>;

  /**
   * Build a dummy JWT with a base64url‐encoded payload.
   * @param payload The JWT payload (must include `exp` when testing setToken).
   * @returns A token string of form "header.payload.signature"
   */
  function createToken(payload: Record<string, unknown>): string {
    const json = JSON.stringify(payload);
    const b64 = btoa(json)
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '');
    return `header.${b64}.signature`;
  }

  beforeEach(() => {
    // Ensure key/iv are exactly 16 chars so AES-128 works
    environment.tokenSecret = '1234567890123456';
    environment.tokenIv = '6543210987654321';

    // Mock document.cookie get/set
    cookieStore = {};
    spyOnProperty(document, 'cookie', 'get').and.callFake(() => {
      return Object.entries(cookieStore)
        .map(([k, v]) => `${k}=${v}`)
        .join('; ');
    });
    spyOnProperty(document, 'cookie', 'set').and.callFake(
      (cookieStr: string) => {
        const [nameValue] = cookieStr.split(';');
        const [name, rawValue] = nameValue.split('=');
        if (!rawValue) {
          delete cookieStore[name];
        } else {
          cookieStore[name] = decodeURIComponent(rawValue);
        }
      }
    );

    jasmine.clock().install();
    service = new SessionService();
  });

  afterEach(() => {
    jasmine.clock().uninstall();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('initially emits false for isAuthenticated$', (done) => {
    service.isAuthenticated$.pipe(take(1)).subscribe((val) => {
      expect(val).toBe(false);
      done();
    });
  });

  it('throws if setting a token without exp claim', () => {
    const nowSec = Math.floor(Date.now() / 1000);
    const tokenNoExp = createToken({ iat: nowSec });
    expect(() => service.setToken(tokenNoExp)).toThrowError(
      'Token JWT sans claim exp'
    );
  });

  it('setToken stores an encrypted cookie and emits true', () => {
    const nowSec = Math.floor(Date.now() / 1000);
    const payload = { exp: nowSec + 60, iat: nowSec, user_id: 'user123' };
    const token = createToken(payload);
    const values: boolean[] = [];
    service.isAuthenticated$.subscribe((v) => values.push(v));

    service.setToken(token);

    // initial false then true
    expect(values).toEqual([false, true]);
    expect(cookieStore['auth_token']).toBeDefined();
  });

  it('getToken returns the original JWT', () => {
    const nowSec = Math.floor(Date.now() / 1000);
    const payload = { exp: nowSec + 60, iat: nowSec, user_id: 'user123' };
    const token = createToken(payload);

    service.setToken(token);
    expect(service.getToken()).toBe(token);
  });

  it('clearSession removes cookie and emits false', () => {
    const nowSec = Math.floor(Date.now() / 1000);
    const payload = { exp: nowSec + 60, iat: nowSec, user_id: 'user123' };
    const token = createToken(payload);
    const values: boolean[] = [];
    service.isAuthenticated$.subscribe((v) => values.push(v));

    service.setToken(token);
    service.clearSession();

    // false → true on setToken → false on clearSession
    expect(values).toEqual([false, true, false]);
    expect(cookieStore['auth_token']).toBeUndefined();
  });

  it('getUserId returns the user_id claim', () => {
    const nowSec = Math.floor(Date.now() / 1000);
    const payload = { exp: nowSec + 60, iat: nowSec, user_id: 'u42' };
    const token = createToken(payload);

    service.setToken(token);
    expect(service.getUserId()).toBe('u42');
  });

  it('getPayload returns the full decoded payload', () => {
    const nowSec = Math.floor(Date.now() / 1000);
    const payload = { exp: nowSec + 120, iat: nowSec, foo: 'bar' };
    const token = createToken(payload);

    service.setToken(token);
    expect(service.getPayload()).toEqual(payload);
  });

  it('getExpirationDate returns the exp claim as Date', () => {
    const nowSec = Math.floor(Date.now() / 1000);
    const payload = { exp: nowSec + 120, iat: nowSec };
    const token = createToken(payload);

    service.setToken(token);
    expect(service.getExpirationDate()).toEqual(
      new Date((nowSec + 120) * 1000)
    );
  });

  it('isTokenExpired is true when no token present', () => {
    expect(service.isTokenExpired()).toBe(true);
  });

  it('isTokenExpired is false for a valid token', () => {
    const nowSec = Math.floor(Date.now() / 1000);
    const payload = { exp: nowSec + 60, iat: nowSec };
    const token = createToken(payload);

    service.setToken(token);
    expect(service.isTokenExpired()).toBe(false);
  });

  it('isTokenExpired is true for an already-expired token', () => {
    const nowSec = Math.floor(Date.now() / 1000);
    const payload = { exp: nowSec - 10, iat: nowSec - 20 };
    const token = createToken(payload);

    service.setToken(token);
    expect(service.isTokenExpired()).toBe(true);
  });

  it('automatically clears session when token expires via timer', () => {
    const start = Date.now();
    const nowSec = Math.floor(start / 1000);
    const payload = { exp: nowSec + 1, iat: nowSec, user_id: 'timed' };
    const token = createToken(payload);

    const values: boolean[] = [];
    service.isAuthenticated$.subscribe((v) => values.push(v));

    service.setToken(token);
    // should have fired true immediately
    expect(values).toEqual([false, true]);

    // advance 1 second → clearSession() via setTimeout
    jasmine.clock().tick(1000);
    expect(values).toEqual([false, true, false]);
    expect(service.isTokenExpired()).toBe(true);
    expect(cookieStore['auth_token']).toBeUndefined();
  });
});
