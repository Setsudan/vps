import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import {
  provideHttpClientTesting,
  HttpTestingController,
} from '@angular/common/http/testing';
import { AuthService } from './auth.service';
import { SessionService } from '../session/session.service';
import { UserStateService } from '../user-state/user-state.service';
import { environment } from '../../../environments/environment';
import { APIResponse } from '../../../types/apiResponse';
import { IUser } from '../../../types/user';
import { of } from 'rxjs';

describe('AuthService', () => {
  let service: AuthService;
  let httpMock: HttpTestingController;
  let sessionService: jasmine.SpyObj<SessionService>;
  let userStateService: jasmine.SpyObj<UserStateService>;
  const apiUrl: string = environment.apiUrl;

  function createToken(payload: Record<string, unknown>): string {
    const json = JSON.stringify(payload);
    const b64 = btoa(json)
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
    userStateService = jasmine.createSpyObj('UserStateService', [
      'set',
      'clear',
    ]);

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        AuthService,
        { provide: SessionService, useValue: sessionService },
        { provide: UserStateService, useValue: userStateService },
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

  it('should POST to /auth/login, store token, fetch user, and set user state', () => {
    const email = 'bob@example.com';
    const password = 'hunter2';
    const fakeToken = createToken({ exp: Math.floor(Date.now() / 1000) + 60 });
    const fakeUser: IUser = {
      id: '123',
      username: 'bob',
      email,
      bio: '',
      role: '',
      avatar: '',
      status: 'online',
      createdAt: new Date(),
      updatedAt: new Date(),
      guilds: [],
    };

    service.login(email, password).subscribe((result) => {
      expect(result).toBeUndefined();
      expect(sessionService.setToken).toHaveBeenCalledWith(fakeToken);
      expect(userStateService.set).toHaveBeenCalledWith(fakeUser);
    });

    // 1) POST /auth/login
    const loginReq = httpMock.expectOne(`${apiUrl}/auth/login`);
    expect(loginReq.request.method).toBe('POST');
    expect(loginReq.request.body).toEqual({ email, password });
    const loginResp: APIResponse<{ token: string }> = {
      data: { token: fakeToken },
      code: 0,
      message: '',
    };
    loginReq.flush(loginResp);

    // 2) GET /user/me
    const meReq = httpMock.expectOne(`${apiUrl}/user/me`);
    expect(meReq.request.method).toBe('GET');
    const meResp: APIResponse<IUser> = {
      data: fakeUser,
      code: 0,
      message: '',
    };
    meReq.flush(meResp);
  });

  it('logout() should clear the session via SessionService', () => {
    service.logout();
    expect(sessionService.clearSession).toHaveBeenCalled();
    expect(userStateService.clear).toHaveBeenCalled();
  });

  it('isAuthenticated$ should proxy SessionService.isAuthenticated$', () => {
    expect(service.isAuthenticated$).toBe(sessionService.isAuthenticated$);
  });
});
