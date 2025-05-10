import { ComponentFixture, TestBed } from '@angular/core/testing';
import { LoginComponent } from './login.component';
import { ReactiveFormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../../../services/auth/auth.service';
import { of, throwError } from 'rxjs';

describe('LoginComponent', () => {
  let component: LoginComponent;
  let fixture: ComponentFixture<LoginComponent>;
  let authService: jasmine.SpyObj<AuthService>;
  let router: jasmine.SpyObj<Router>;

  beforeEach(async () => {
    // create spies
    authService = jasmine.createSpyObj('AuthService', ['login'], {
      isAuthenticated$: of(false),
    });
    router = jasmine.createSpyObj('Router', ['navigate']);

    await TestBed.configureTestingModule({
      imports: [LoginComponent, ReactiveFormsModule],
      providers: [
        { provide: AuthService, useValue: authService },
        { provide: Router, useValue: router },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(LoginComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create the component', () => {
    expect(component).toBeTruthy();
  });

  it('should have form invalid initially', () => {
    expect(component.form.invalid).toBeTrue();
    expect(component.loading).toBeFalse();
    expect(component.error).toBeNull();
  });

  it('should mark all controls touched and not call login if form is invalid on submit', () => {
    spyOn(component.form, 'markAllAsTouched');
    component.onSubmit();
    expect(component.form.markAllAsTouched).toHaveBeenCalled();
    expect(authService.login).not.toHaveBeenCalled();
    expect(component.loading).toBeFalse();
  });

  it('should call AuthService.login and navigate on successful submit', () => {
    component.form.setValue({ email: 'a@x.com', password: 'pass' });
    authService.login.and.returnValue(of(void 0));

    component.onSubmit();

    expect(authService.login).toHaveBeenCalledWith('a@x.com', 'pass');
    expect(component.loading).toBeFalse();
    expect(component.error).toBeNull();
    expect(router.navigate).toHaveBeenCalledWith(['']);
  });

  it('should set error message and reset loading on login error', () => {
    const err = new Error('Invalid credentials');
    component.form.setValue({ email: 'fail@x.com', password: 'bad' });
    authService.login.and.returnValue(throwError(() => err));
    spyOn(console, 'error');

    component.onSubmit();

    expect(authService.login).toHaveBeenCalledWith('fail@x.com', 'bad');
    expect(console.error).toHaveBeenCalledWith(err);
    expect(component.error).toBe('Invalid credentials');
    expect(component.loading).toBeFalse();
  });
});
