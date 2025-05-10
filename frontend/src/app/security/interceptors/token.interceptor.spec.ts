import { TestBed } from '@angular/core/testing';
import {
  HttpInterceptorFn,
  HttpRequest,
  HttpHandlerFn,
} from '@angular/common/http';
import { of } from 'rxjs';
import { SessionService } from '../../services/session/session.service';
import { tokenInterceptor } from './token.interceptor';

describe('tokenInterceptor', () => {
  let sessionServiceSpy: jasmine.SpyObj<SessionService>;
  const interceptor: HttpInterceptorFn = (req, next) =>
    TestBed.runInInjectionContext(() => tokenInterceptor(req, next));

  beforeEach(() => {
    sessionServiceSpy = jasmine.createSpyObj('SessionService', ['getToken']);
    TestBed.configureTestingModule({
      providers: [{ provide: SessionService, useValue: sessionServiceSpy }],
    });
  });

  it('should add Authorization header if token exists', () => {
    sessionServiceSpy.getToken.and.returnValue('test-token');
    const mockRequest = new HttpRequest('GET', '/test');
    const mockHandler: HttpHandlerFn = (req) => {
      expect(req.headers.get('Authorization')).toBe('Bearer test-token');
      return of(null as any);
    };

    interceptor(mockRequest, mockHandler);
  });

  it('should not add Authorization header if token does not exist', () => {
    sessionServiceSpy.getToken.and.returnValue(null);
    const mockRequest = new HttpRequest('GET', '/test');
    const mockHandler: HttpHandlerFn = (req) => {
      expect(req.headers.has('Authorization')).toBeFalse();
      return of(null as any);
    };

    interceptor(mockRequest, mockHandler);
  });

  it('should set withCredentials to true', () => {
    sessionServiceSpy.getToken.and.returnValue('test-token');
    const mockRequest = new HttpRequest('GET', '/test');
    const mockHandler: HttpHandlerFn = (req) => {
      expect(req.withCredentials).toBeTrue();
      return of(null as any);
    };

    interceptor(mockRequest, mockHandler);
  });
});
