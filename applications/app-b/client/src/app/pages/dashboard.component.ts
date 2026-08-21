import { CommonModule } from '@angular/common';
import { Component, OnDestroy, OnInit, inject } from '@angular/core';
import { Subscription, interval } from 'rxjs';
import { AuthService, errorMessage } from '../services/auth.service';

@Component({
  selector: 'app-dashboard',
  imports: [CommonModule],
  template: `
    <div class="app-shell">
      <header class="topbar"><div class="brand"><span class="crest">B</span><span><b>Bolt</b><small>Trusted by Scaean Gate</small></span></div>
        @if (authenticated) { <div class="actions"><button (click)="localLogout()">Local logout</button><button class="danger" (click)="ssoLogout()">End SSO session</button></div> }
      </header>
      @if (notice) { <div class="notice">{{ notice }}</div> }
      <main>
        @if (loading) { <section class="center"><span class="spinner"></span><p>Checking the local seal…</p></section> }
        @else if (ended) { <section class="empty-card ended"><p class="eyebrow">Session ended</p><h1>Your access was revoked</h1><p>{{ endReason || 'The central identity provider ended this local session.' }}</p><button class="primary" (click)="login()">Start a new session</button></section> }
        @else if (!authenticated) { <section class="empty-card"><span class="large-crest">B</span><p class="eyebrow">Bolt workspace</p><h1>Enter with one identity</h1><p>Authenticate through Scaean Gate. Bolt never receives or stores your password.</p><button class="primary" (click)="login()">Continue with Scaean Gate</button></section> }
        @else {
          <div class="dashboard">
            <section class="paper-card identity"><p class="eyebrow">Identity</p><div class="identity-row"><span class="avatar">{{ user.name?.charAt(0) }}</span><div><h1>Hello, {{ user.name }}</h1><p>{{ user.email }}</p><span class="tag" *ngFor="let group of user.groups">{{ group }}</span></div></div></section>
            <section class="paper-card session"><div class="section-head"><div><p class="eyebrow">Local session</p><h2>Bolt session</h2></div><span class="status active">● {{ session.status }}</span></div><dl><div><dt>Session ID</dt><dd><code>{{ session.id }}</code></dd></div><div><dt>Created</dt><dd>{{ session.createdAt | date:'medium' }}</dd></div><div><dt>Expires</dt><dd>{{ session.expiresAt | date:'medium' }}</dd></div></dl></section>
            <section class="paper-card wide"><div class="section-head"><div><p class="eyebrow">Verified by the backend</p><h2>Authentication timeline</h2></div><button (click)="loadDetails()">Refresh</button></div>
              <div class="timeline">@for (item of activities; track item.id; let i=$index) { <div class="timeline-item"><span class="step">{{ i + 1 }}</span><div><b>{{ label(item.event_type) }}</b><small>{{ item.created_at | date:'medium' }}</small></div><span class="status active">{{ item.result }}</span></div> } @empty { <p class="muted">No activity recorded for this session.</p> }</div>
            </section>
            <section class="paper-card wide"><div class="section-head"><div><p class="eyebrow">Back-channel sync</p><h2>Processed events</h2></div><button (click)="loadDetails()">Refresh</button></div>
              <div class="table-wrap"><table><thead><tr><th>Processed</th><th>Event</th><th>Event ID</th><th>Result</th></tr></thead><tbody>@for (event of events; track event.event_id) { <tr><td>{{ event.processed_at | date:'medium' }}</td><td><b>{{ event.event_type }}</b></td><td><code>{{ event.event_id }}</code></td><td><span class="status active">{{ event.result }}</span></td></tr> } @empty { <tr><td colspan="4" class="muted">No revocation events have been processed.</td></tr> }</tbody></table></div>
            </section>
          </div>
        }
      </main>
    </div>
  `,
})
export class DashboardComponent implements OnInit, OnDestroy {
  private auth = inject(AuthService); private polling?: Subscription;
  loading=true; authenticated=false; ended=false; endReason=''; notice=''; user:any={groups:[]};session:any={};activities:any[]=[];events:any[]=[];
  ngOnInit(){this.check();this.polling=interval(5000).subscribe(()=>this.check(true));}
  ngOnDestroy(){this.polling?.unsubscribe();}
  check(silent=false){if(!silent)this.loading=true;this.auth.me().subscribe({next:r=>{this.authenticated=true;this.ended=false;this.user=r.user;this.session=r.session;this.loading=false;this.loadDetails();},error:()=>{this.authenticated=false;this.auth.status().subscribe({next:r=>{this.ended=r.status==='revoked'||r.status==='expired';this.endReason=r.reason||'';this.loading=false;},error:()=>{this.ended=false;this.loading=false;}});}});}
  loadDetails(){if(!this.authenticated)return;this.auth.activity().subscribe({next:r=>this.activities=r||[]});this.auth.events().subscribe({next:r=>this.events=r||[]});}
  label(type:string){return ({AUTH_REDIRECT:'Redirected to Scaean Gate',AUTH_CODE_RECEIVED:'Authorization code received',TOKEN_EXCHANGED:'Access token exchanged',LOCAL_SESSION_CREATED:'Local session created'} as any)[type]||type;}
  login(){this.auth.login();}
  localLogout(){if(!confirm('End only this Bolt session?'))return;this.auth.localLogout().subscribe({next:()=>{this.authenticated=false;this.ended=false;this.say('Local session ended.');},error:e=>this.say(errorMessage(e))});}
  ssoLogout(){if(!confirm('End the central SSO session and revoke access in every connected app?'))return;this.auth.ssoLogout().subscribe({next:()=>{this.say('Central session ended. Waiting for revocation…');window.setTimeout(()=>this.check(true),1200);},error:e=>this.say(errorMessage(e))});}
  say(message:string){this.notice=message;window.setTimeout(()=>this.notice='',3500);}
}
