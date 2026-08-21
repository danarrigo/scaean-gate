import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="login-container">
      <div class="login-card">
        <div class="header">
          <div class="logo">
            <span class="gate-icon">🏛️</span>
            <h1>Scaean Gate</h1>
          </div>
          <p class="subtitle">Central Single Sign-On Provider</p>
        </div>

        @if (errorMessage) {
          <div class="alert alert-error">
            <span>⚠️ {{ errorMessage }}</span>
          </div>
        }

        <form (ngSubmit)="onSubmit()">
          <div class="form-group">
            <label for="email">Email Address</label>
            <input
              type="email"
              id="email"
              name="email"
              [(ngModel)]="email"
              placeholder="admin@scaean-gate.com"
              required
            />
          </div>

          <div class="form-group">
            <label for="password">Password</label>
            <input
              type="password"
              id="password"
              name="password"
              [(ngModel)]="password"
              placeholder="••••••••••••"
              required
            />
          </div>

          <button type="submit" class="btn-primary" [disabled]="loading">
            @if (loading) {
              <span>Signing in...</span>
            } @else {
              <span>Sign In</span>
            }
          </button>
        </form>

        <div class="footer">
          <p>Demo Accounts:</p>
          <div class="credentials">
            <code>admin&#64;scaean-gate.com / password123</code>
            <code>testuser&#64;scaean-gate.com / password123</code>
          </div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .login-container {
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 20px;
      background: linear-gradient(135deg, #0f172a 0%, #1e1b4b 50%, #311042 100%);
    }

    .login-card {
      width: 100%;
      max-width: 420px;
      background: rgba(255, 255, 255, 0.95);
      backdrop-filter: blur(10px);
      border-radius: 16px;
      padding: 36px;
      box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.35);
    }

    .header {
      text-align: center;
      margin-bottom: 28px;
    }

    .logo {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 10px;
      margin-bottom: 6px;
    }

    .gate-icon {
      font-size: 2rem;
    }

    h1 {
      font-size: 1.6rem;
      font-weight: 700;
      color: #0f172a;
    }

    .subtitle {
      color: #64748b;
      font-size: 0.9rem;
    }

    .alert {
      padding: 12px 14px;
      border-radius: 8px;
      font-size: 0.875rem;
      margin-bottom: 20px;
    }

    .alert-error {
      background-color: #fee2e2;
      color: #991b1b;
      border: 1px solid #fecaca;
    }

    .form-group {
      margin-bottom: 20px;
    }

    label {
      display: block;
      font-size: 0.875rem;
      font-weight: 500;
      color: #334155;
      margin-bottom: 6px;
    }

    input {
      width: 100%;
      padding: 11px 14px;
      border: 1.5px solid #cbd5e1;
      border-radius: 8px;
      font-size: 0.95rem;
      font-family: inherit;
      outline: none;
      transition: border-color 0.15s ease;
    }

    input:focus {
      border-color: #4f46e5;
      box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.15);
    }

    .btn-primary {
      width: 100%;
      padding: 12px;
      background: #4f46e5;
      color: white;
      border: none;
      border-radius: 8px;
      font-size: 1rem;
      font-weight: 600;
      transition: background 0.15s ease;
    }

    .btn-primary:hover:not(:disabled) {
      background: #4338ca;
    }

    .btn-primary:disabled {
      opacity: 0.6;
      cursor: not-allowed;
    }

    .footer {
      margin-top: 28px;
      padding-top: 20px;
      border-top: 1px solid #e2e8f0;
      font-size: 0.8rem;
      color: #64748b;
      text-align: center;
    }

    .credentials {
      margin-top: 8px;
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    code {
      background: #f1f5f9;
      padding: 3px 6px;
      border-radius: 4px;
      color: #334155;
      font-size: 0.75rem;
    }
  `]
})
export class LoginComponent {
  private api = inject(ApiService);
  private router = inject(Router);
  private route = inject(ActivatedRoute);

  email = 'admin@scaean-gate.com';
  password = 'password123';
  loading = false;
  errorMessage = '';

  onSubmit() {
    this.loading = true;
    this.errorMessage = '';

    this.api.login({ email: this.email, password: this.password }).subscribe({
      next: () => {
        this.loading = false;
        const returnTo = this.route.snapshot.queryParams['return_to'];
        if (returnTo) {
          window.location.href = `http://localhost:8080${returnTo}`;
        } else {
          this.router.navigate(['/admin']);
        }
      },
      error: (err) => {
        this.loading = false;
        this.errorMessage = err.error?.error?.message || 'Login failed. Please check your credentials.';
      }
    });
  }
}
