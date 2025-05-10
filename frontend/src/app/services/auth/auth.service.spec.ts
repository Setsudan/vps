import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import {
  provideHttpClientTesting,
  HttpTestingController,
} from '@angular/common/http/testing';
import { AuthService } from './auth.service';
import { SessionService } from '../session/session.service';
import { environment } from '../../../environments/environment';
import { APIResponse } from '../../../types/apiResponse';
import { of } from 'rxjs';

describe('AuthService', () => {
  let service: AuthService;
  let httpMock: HttpTestingController;
  let sessionService: jasmine.SpyObj<SessionService>;
  const apiUrl: string = environment.apiUrl;

  function createToken(payload: Record<string, unknown>): string {
    const json: string = JSON.stringify(payload);
    const b64: string = btoa(json)
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '');
    return `header.${b64}.signature`;
  }

  beforeEach(() => {
    sessionService = jasmine.createSpyObj(
      'SessionService',
      ['setToken', 'clearSession'],
      { isAuthenticated$: of(false) }
    );

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        AuthService,
        { provide: SessionService, useValue: sessionService },
      ],
    });

    service = TestBed.inject(AuthService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it('should POST to /auth/register and complete without error', () => {
    const payload = {
      username: 'alice',
      email: 'alice@example.com',
      password: 'secret',
    };
    service.register(payload).subscribe((result) => {
      expect(result).toBeNull();
    });

    const req = httpMock.expectOne(`${apiUrl}/auth/register`);
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual(payload);
    req.flush(null, { status: 201, statusText: 'Created' });
  });

  it('should POST to /auth/login, store token, and complete', () => {
    const email = 'bob@example.com';
    const password = 'hunter2';
    const fakeToken = createToken({ exp: Math.floor(Date.now() / 1000) + 60 });

    service.login(email, password).subscribe((result) => {
      expect(result).toBeUndefined();
      expect(sessionService.setToken).toHaveBeenCalledWith(fakeToken);
    });

    const req = httpMock.expectOne(`${apiUrl}/auth/login`);
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ email, password });

    const apiResp: APIResponse<{ token: string }> = {
      data: { token: fakeToken },
      code: 0,
      message: '',
    };
    req.flush(apiResp);
  });

  it('logout() should clear the session via SessionService', () => {
    service.logout();
    expect(sessionService.clearSession).toHaveBeenCalled();
  });

  it('isAuthenticated$ should proxy SessionService.isAuthenticated$', () => {
    expect(service.isAuthenticated$).toBe(sessionService.isAuthenticated$);
  });
});
