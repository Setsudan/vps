import { TestBed } from '@angular/core/testing';
import {
  Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot,
} from '@angular/router';
import { authGuard } from './auth.guard';
import { SessionService } from '../../services/session/session.service';

describe('authGuard', () => {
  let routerSpy: jasmine.SpyObj<Router>;
  let sessionSpy: jasmine.SpyObj<SessionService>;
  const dummyRoute = {} as ActivatedRouteSnapshot;
  const dummyState = { url: '/test' } as RouterStateSnapshot;

  beforeEach(() => {
    // create jasmine spies
    routerSpy = jasmine.createSpyObj('Router', ['navigate']);
    sessionSpy = jasmine.createSpyObj('SessionService', ['isTokenExpired']);

    TestBed.configureTestingModule({
      providers: [
        // cast to satisfy the injector
        { provide: Router, useValue: routerSpy as unknown as Router },
        {
          provide: SessionService,
          useValue: sessionSpy as unknown as SessionService,
        },
      ],
    });
  });

  /**
   * Runs the authGuard inside Angular's DI context.
   * @returns true if activation allowed, false otherwise.
   */
  function runGuard(): boolean {
    return TestBed.runInInjectionContext(
      () => authGuard(dummyRoute, dummyState) as boolean
    );
  }

  it('should redirect to ["auth","login"] and return false when token expired', () => {
    sessionSpy.isTokenExpired.and.returnValue(true);

    const result = runGuard();

    expect(routerSpy.navigate).toHaveBeenCalledWith(['auth', 'login']);
    expect(result).toBe(false);
  });

  it('should return true and not navigate when token not expired', () => {
    sessionSpy.isTokenExpired.and.returnValue(false);

    const result = runGuard();

    expect(routerSpy.navigate).not.toHaveBeenCalled();
    expect(result).toBe(true);
  });
});
