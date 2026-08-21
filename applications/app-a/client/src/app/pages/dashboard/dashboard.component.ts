import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AuthService } from '../../services/auth.service';
import { Subscription, interval } from 'rxjs';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="app-container">
      <!-- Top Navbar -->
      <header class="navbar">
        <div class="brand">
          <span class="brand-icon">🅰️</span>
          <div>
            <h2>Apex Workspace</h2>
            <span class="sub-brand">Relying Application A (Port 4201)</span>
          </div>
        </div>

        @if (authenticated) {
          <div class="nav-actions">
            <button class="btn-sm btn-outline" (click)="onLocalLogout()">🚪 Local Logout</button>
            <button class="btn-sm btn-danger" (click)="onSSOLogout()">🔴 SSO Global Logout</button>
          </div>
        }
      </header>

      <!-- Main Body -->
      <main class="main-content">
        @if (loading) {
          <div class="loading-box">
            <div class="spinner"></div>
            <p>Verifying session with Apex backend...</p>
          </div>
        } @else if (!authenticated) {
          <!-- Unauthenticated Landing State -->
          <div class="login-box">
            <div class="login-card">
              <span class="large-icon">🔒</span>
              <h3>Authentication Required</h3>
              <p class="desc">You must sign in via the Central SSO Provider to access Apex Workspace.</p>
              
              <button class="btn-primary btn-lg" (click)="login()">
                <span>🔑 Sign In with Single Sign-On</span>
              </button>

              <div class="note">
                <span>Directs to Central SSO Server (Port 8080)</span>
              </div>
            </div>
          </div>
        } @else {
          <!-- Authenticated Dashboard View -->
          <div class="dashboard-grid">
            <!-- 1. Greeting & Profile Card -->
            <div class="card greeting-card">
              <div class="card-header">
                <span class="section-badge">User Identity</span>
              </div>
              <div class="greeting-body">
                <div class="avatar">{{ user.name.charAt(0) }}</div>
                <div>
                  <h3>Hello, {{ user.name }}!</h3>
                  <p class="email">{{ user.email }}</p>
                  <div class="groups-list">
                    @for (g of user.groups; track g) {
                      <span class="group-tag">🛡️ {{ g }}</span>
                    }
                  </div>
                </div>
              </div>
            </div>

            <!-- 2. Session Status Banner -->
            <div class="card status-card">
              <div class="card-header">
                <span class="section-badge">Local Session State</span>
                <span class="badge" 
                  [class.badge-active]="session.status === 'active'"
                  [class.badge-revoked]="session.status === 'revoked'"
                  [class.badge-expired]="session.status === 'expired'">
                  ● {{ session.status | uppercase }}
                </span>
              </div>

              <div class="session-details">
                <div class="detail-row">
                  <span class="label">Session ID:</span>
                  <code>{{ session.id }}</code>
                </div>
                <div class="detail-row">
                  <span class="label">Created:</span>
                  <span>{{ session.createdAt | date:'medium' }}</span>
                </div>
                <div class="detail-row">
                  <span class="label">Expires:</span>
                  <span>{{ session.expiresAt | date:'medium' }}</span>
                </div>
              </div>

              @if (session.status === 'revoked') {
                <div class="revoked-banner">
                  ⚠️ This session was revoked via Back-Channel SSO Logout!
                </div>
              }
            </div>

            <!-- 3. Processed Events Log (Back-Channel Webhook Tracker) -->
            <div class="card full-width">
              <div class="card-header">
                <div>
                  <h3>Processed Events Log</h3>
                  <p class="subtitle">Back-channel synchronization events received from Sync Worker</p>
                </div>
                <button class="btn-xs btn-outline" (click)="loadEvents()">🔄 Refresh</button>
              </div>

              @if (events.length === 0) {
                <div class="empty-state">
                  <span>No back-channel events recorded yet. Try logging out on Central SSO!</span>
                </div>
              } @else {
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>Event ID</th>
                      <th>Event Type</th>
                      <th>Processed At</th>
                      <th>Result</th>
                    </tr>
                  </thead>
                  <tbody>
                    @for (e of events; track e.eventId) {
                      <tr>
                        <td><code>{{ e.eventId }}</code></td>
                        <td><strong>{{ e.eventType }}</strong></td>
                        <td class="text-muted">{{ e.processedAt | date:'medium' }}</td>
                        <td>
                          <span class="badge badge-active">{{ e.result }}</span>
                        </td>
                      </tr>
                    }
                  </tbody>
                </table>
              }
            </div>
          </div>
        }
      </main>
    </div>
  `,
  styles: [`
    .app-container {
      min-height: 100vh;
      display: flex;
      flex-direction: column;
      background: #f8fafc;
    }

    .navbar {
      background: white;
      border-bottom: 1px solid #e2e8f0;
      padding: 16px 36px;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    .brand {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .brand-icon {
      font-size: 2rem;
    }

    .brand h2 {
      font-size: 1.25rem;
      font-weight: 700;
      color: #0f172a;
    }

    .sub-brand {
      font-size: 0.8rem;
      color: #64748b;
    }

    .nav-actions {
      display: flex;
      gap: 12px;
    }

    .main-content {
      flex: 1;
      padding: 36px;
      max-width: 1100px;
      margin: 0 auto;
      width: 100%;
    }

    .loading-box {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 80px 0;
      color: #64748b;
      gap: 16px;
    }

    .spinner {
      width: 40px;
      height: 40px;
      border: 3px solid #e2e8f0;
      border-top-color: #0284c7;
      border-radius: 50%;
      animation: spin 0.8s linear infinite;
    }

    @keyframes spin {
      to { transform: rotate(360deg); }
    }

    .login-box {
      display: flex;
      justify-content: center;
      padding-top: 60px;
    }

    .login-card {
      background: white;
      border: 1px solid #e2e8f0;
      border-radius: 16px;
      padding: 40px;
      max-width: 460px;
      width: 100%;
      text-align: center;
      box-shadow: 0 10px 15px -3px rgba(0,0,0,0.05);
    }

    .large-icon {
      font-size: 3rem;
      margin-bottom: 16px;
      display: inline-block;
    }

    .login-card h3 {
      font-size: 1.35rem;
      font-weight: 700;
      color: #0f172a;
      margin-bottom: 8px;
    }

    .desc {
      color: #64748b;
      font-size: 0.95rem;
      margin-bottom: 28px;
    }

    .btn-lg {
      width: 100%;
      padding: 14px;
      font-size: 1rem;
      font-weight: 600;
    }

    .note {
      margin-top: 18px;
      font-size: 0.8rem;
      color: #94a3b8;
    }

    .dashboard-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 24px;
    }

    .full-width {
      grid-column: 1 / -1;
    }

    .card {
      background: white;
      border: 1px solid #e2e8f0;
      border-radius: 12px;
      padding: 24px;
      box-shadow: 0 1px 3px rgba(0,0,0,0.04);
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 18px;
    }

    .section-badge {
      font-size: 0.75rem;
      font-weight: 700;
      text-transform: uppercase;
      color: #64748b;
      letter-spacing: 0.5px;
    }

    .greeting-body {
      display: flex;
      align-items: center;
      gap: 16px;
    }

    .avatar {
      width: 56px;
      height: 56px;
      border-radius: 50%;
      background: #0284c7;
      color: white;
      font-size: 1.5rem;
      font-weight: 700;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .greeting-body h3 {
      font-size: 1.2rem;
      font-weight: 700;
      color: #0f172a;
    }

    .email {
      color: #64748b;
      font-size: 0.9rem;
      margin-bottom: 8px;
    }

    .groups-list {
      display: flex;
      gap: 6px;
      flex-wrap: wrap;
    }

    .group-tag {
      background: #f1f5f9;
      color: #334155;
      padding: 2px 8px;
      border-radius: 6px;
      font-size: 0.75rem;
      font-weight: 600;
    }

    .session-details {
      display: flex;
      flex-direction: column;
      gap: 10px;
      font-size: 0.875rem;
    }

    .detail-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .label {
      color: #64748b;
      font-weight: 500;
    }

    code {
      background: #f1f5f9;
      padding: 2px 6px;
      border-radius: 4px;
      font-size: 0.75rem;
    }

    .revoked-banner {
      margin-top: 14px;
      padding: 10px;
      background: #fee2e2;
      color: #991b1b;
      border-radius: 6px;
      font-size: 0.85rem;
      font-weight: 600;
    }

    .subtitle {
      font-size: 0.85rem;
      color: #64748b;
    }

    .empty-state {
      padding: 32px;
      text-align: center;
      color: #94a3b8;
      font-size: 0.9rem;
    }

    .data-table {
      width: 100%;
      border-collapse: collapse;
      font-size: 0.875rem;
      margin-top: 12px;
    }

    .data-table th {
      text-align: left;
      padding: 10px 12px;
      background: #f8fafc;
      color: #64748b;
      font-weight: 600;
      border-bottom: 1px solid #e2e8f0;
    }

    .data-table td {
      padding: 12px;
      border-bottom: 1px solid #f1f5f9;
    }

    .btn-sm {
      padding: 8px 14px;
      border-radius: 6px;
      font-size: 0.85rem;
      font-weight: 600;
      border: none;
    }

    .btn-xs {
      padding: 4px 10px;
      border-radius: 4px;
      font-size: 0.75rem;
      font-weight: 600;
      border: 1px solid #cbd5e1;
      background: white;
    }

    .btn-primary { background: #0284c7; color: white; border: none; border-radius: 8px; }
    .btn-primary:hover { background: #0369a1; }
    .btn-outline { background: white; border: 1px solid #cbd5e1; color: #334155; }
    .btn-outline:hover { background: #f8fafc; }
    .btn-danger { background: #ef4444; color: white; }
    .btn-danger:hover { background: #dc2626; }
  `]
})
export class DashboardComponent implements OnInit, OnDestroy {
  private auth = inject(AuthService);
  private pollSub?: Subscription;

  loading = true;
  authenticated = false;

  user: any = { name: '', email: '', groups: [] };
  session: any = { id: '', status: '', createdAt: '', expiresAt: '' };
  events: any[] = [];

  ngOnInit() {
    this.checkSession();
    // Live polling every 3 seconds to detect back-channel logouts in real-time!
    this.pollSub = interval(3000).subscribe(() => {
      if (this.authenticated) {
        this.checkSession(true);
      }
    });
  }

  ngOnDestroy() {
    this.pollSub?.unsubscribe();
  }

  checkSession(silent = false) {
    if (!silent) this.loading = true;

    this.auth.getMe().subscribe({
      next: (res) => {
        this.authenticated = true;
        this.user = res.user;
        this.session = res.session;
        this.loading = false;
        this.loadEvents();
      },
      error: () => {
        this.authenticated = false;
        this.loading = false;
      }
    });

    // Fallback if backend is offline
    setTimeout(() => {
      if (this.loading) {
        this.loading = false;
        this.authenticated = false;
      }
    }, 1500);
  }

  loadEvents() {
    this.auth.getEvents().subscribe({
      next: (res) => this.events = res || [],
      error: () => this.events = []
    });
  }

  login() {
    this.auth.redirectToLogin();
  }

  onLocalLogout() {
    this.auth.logout().subscribe(() => {
      this.authenticated = false;
      this.checkSession();
    });
  }

  onSSOLogout() {
    this.auth.redirectToSSOLogout();
  }
}
