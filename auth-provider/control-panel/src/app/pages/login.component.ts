import { Component, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { ApiService, errorMessage } from '../services/api.service';

@Component({
  selector: 'app-login',
  imports: [FormsModule, RouterLink],
  template: `
    <main class="auth-page">
      <div class="auth-frame">
        <figure class="login-art">
          <img src="/scaean-gate-login.png" alt="An ancient fortified gate guarded by an assembled army" />
        </figure>
        <section class="auth-card">
          <div class="sigil" aria-hidden="true"><span>SG</span></div>
          <p class="eyebrow">Central identity</p>
          <h1>Enter Scaean Gate</h1>
          <p class="muted">One identity for every trusted realm.</p>
          <form (ngSubmit)="submit()">
            <label>Email<input name="email" type="email" [(ngModel)]="email" autocomplete="username" required /></label>
            <label>Password<input name="password" type="password" [(ngModel)]="password" autocomplete="current-password" required /></label>
            @if (error) { <p class="form-error" role="alert">{{ error }}</p> }
            <button class="primary full" [disabled]="busy">{{ busy ? 'Opening gate…' : 'Sign in' }}</button>
          </form>
          <a routerLink="/" class="quiet-link">Return to account</a>
        </section>
      </div>
    </main>
  `,
})
export class LoginComponent {
  private api = inject(ApiService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  email = '';
  password = '';
  error = '';
  busy = false;

  submit() {
    if (this.busy) return;
    this.busy = true;
    this.error = '';
    this.api.login({ email: this.email.trim(), password: this.password }).subscribe({
      next: () => {
        const returnTo = this.route.snapshot.queryParamMap.get('return_to');
        if (returnTo?.startsWith('/authorize?')) {
          window.location.assign(`${this.api.base}${returnTo}`);
        } else {
          void this.router.navigateByUrl('/');
        }
      },
      error: (error) => { this.error = errorMessage(error); this.busy = false; },
    });
  }
}
