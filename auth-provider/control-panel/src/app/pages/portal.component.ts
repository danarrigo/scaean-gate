import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { ApiService, errorMessage } from '../services/api.service';

type Tab = 'users' | 'groups' | 'apps' | 'policies' | 'audit' | 'events';

@Component({
  selector: 'app-portal',
  imports: [CommonModule, FormsModule],
  template: `
    @if (loading) { <main class="center-state"><span class="spinner"></span><p>Reading the rolls…</p></main> }
    @else if (!profile) { <main class="center-state"><h1>Scaean Gate</h1><p>Your central session is not active.</p><button class="primary" (click)="goLogin()">Sign in</button></main> }
    @else {
      <div class="shell">
        <header class="topbar">
          <a class="brand" href="/"><span class="brand-mark">SG</span><span><b>Scaean Gate</b><small>Identity registry</small></span></a>
          <div class="account"><span>{{ profile.name }}</span><button class="text-button" (click)="logout()">Sign out</button></div>
        </header>

        @if (notice) { <div class="notice" role="status">{{ notice }}</div> }
        @if (!isAdmin) {
          <main class="account-page">
            <section class="paper-card account-card">
              <p class="eyebrow">Central account</p><h1>Welcome, {{ profile.name }}</h1>
              <dl><div><dt>Email</dt><dd>{{ profile.email }}</dd></div><div><dt>Groups</dt><dd><span class="tag" *ngFor="let group of profile.groups">{{ group }}</span></dd></div><div><dt>Session</dt><dd><span class="status active">Active</span></dd></div></dl>
              <button class="danger-outline" (click)="logout()">End central session</button>
            </section>
          </main>
        } @else {
          <div class="admin-layout">
            <aside class="sidebar">
              <p class="eyebrow">Stewardship</p>
              @for (item of nav; track item.id) { <button [class.active]="tab === item.id" (click)="select(item.id)"><span>{{ item.number }}</span>{{ item.label }}</button> }
            </aside>
            <main class="content">
              <div class="page-heading"><div><p class="eyebrow">Administration</p><h1>{{ title }}</h1></div><input class="search" [(ngModel)]="search" placeholder="Search this register" /></div>

              @if (tab === 'users') {
                <section class="paper-card"><div class="section-head"><h2>Users</h2><button class="primary small" (click)="openUser()">Add user</button></div>
                  <div class="table-wrap"><table><thead><tr><th>Name</th><th>Email</th><th>Groups</th><th>Status</th><th></th></tr></thead><tbody>
                    @for (u of filtered(users); track u.id) { <tr><td><b>{{ u.name }}</b></td><td>{{ u.email }}</td><td><span class="tag" *ngFor="let g of u.groups">{{ g.name }}</span></td><td><span class="status" [class.active]="u.status === 'active'">{{ u.status }}</span></td><td class="actions"><button (click)="openUser(u)">Edit</button><button (click)="toggleUser(u)">{{ u.status === 'active' ? 'Deactivate' : 'Activate' }}</button><button class="danger" (click)="deleteUser(u)">Delete</button></td></tr> }
                  </tbody></table></div>
                </section>
              }

              @if (tab === 'groups') {
                <section class="paper-card"><div class="section-head"><h2>Groups</h2><button class="primary small" (click)="openGroup()">Add group</button></div>
                  <div class="card-grid">@for (g of filtered(groups); track g.id) { <article class="registry-card"><h3>{{ g.name }}</h3><p>{{ g.description || 'No description' }}</p><div class="actions"><button (click)="manageGroup(g)">Members</button><button (click)="openGroup(g)">Edit</button><button class="danger" (click)="deleteGroup(g)">Delete</button></div></article> }</div>
                </section>
              }

              @if (tab === 'apps') {
                <section class="paper-card"><div class="section-head"><h2>Applications</h2><button class="primary small" (click)="openApp()">Register app</button></div>
                  <div class="table-wrap"><table><thead><tr><th>Name</th><th>Client</th><th>Secret</th><th>URLs</th><th>Status</th><th></th></tr></thead><tbody>
                    @for (a of filtered(apps); track a.id) { <tr><td><b>{{ a.name }}</b></td><td><code>{{ a.client_id }}</code></td><td><code>{{ a.client_secret_prefix || '—' }}…</code></td><td><a [href]="a.launch_url" target="_blank">Launch</a><small class="block">{{ a.redirect_uris?.length || 0 }} redirect URI(s)</small></td><td><span class="status" [class.active]="a.status === 'active'">{{ a.status }}</span></td><td class="actions"><button (click)="manageURIs(a)">Redirects</button><button (click)="openApp(a)">Edit</button><button class="danger" (click)="deleteApp(a)">Delete</button></td></tr> }
                  </tbody></table></div>
                </section>
              }

              @if (tab === 'policies') {
                <section class="paper-card"><div class="section-head"><h2>Access policies</h2><button class="primary small" (click)="showPolicy = true">Add policy</button></div>
                  <div class="table-wrap"><table><thead><tr><th>Group</th><th>Application</th><th>Effect</th><th></th></tr></thead><tbody>
                    @for (p of filtered(policies); track p.id) { <tr><td>{{ p.group_name }}</td><td>{{ p.application_name }}</td><td><span class="status active">{{ p.effect }}</span></td><td class="actions"><button class="danger" (click)="deletePolicy(p)">Revoke</button></td></tr> }
                  </tbody></table></div>
                </section>
              }

              @if (tab === 'audit') {
                <section class="paper-card"><div class="section-head"><h2>Security audit</h2><select [(ngModel)]="statusFilter"><option value="">All results</option><option>success</option><option>failed</option></select></div>
                  <div class="table-wrap"><table><thead><tr><th>Time</th><th>Event</th><th>Result</th><th>IP address</th><th>Request</th></tr></thead><tbody>
                    @for (log of filteredAudit(); track log.id) { <tr><td>{{ log.created_at | date:'medium' }}</td><td><b>{{ log.event_type }}</b></td><td><span class="status" [class.active]="log.result === 'success'">{{ log.result }}</span></td><td><code>{{ log.ip_address || '—' }}</code></td><td><small>{{ metadata(log.metadata) }}</small></td></tr> }
                  </tbody></table></div>
                </section>
              }

              @if (tab === 'events') {
                <section class="paper-card"><div class="section-head"><h2>Event deliveries</h2><button (click)="load('events')">Refresh</button></div>
                  <div class="table-wrap"><table><thead><tr><th>Created</th><th>Event</th><th>Application</th><th>Status</th><th>Attempts</th><th>Detail</th></tr></thead><tbody>
                    @for (event of filtered(events); track event.id) {
                      @if (event.deliveries?.length) { @for (d of event.deliveries; track d.id) { <tr><td>{{ event.created_at | date:'medium' }}</td><td><b>{{ event.event_type }}</b></td><td>{{ d.application_name }}</td><td><span class="status" [class.active]="d.status === 'succeeded'">{{ d.status }}</span></td><td>{{ d.attempt_count }}/5</td><td><small>{{ d.last_error || (d.processed_at | date:'short') || '—' }}</small></td></tr> } }
                      @else { <tr><td>{{ event.created_at | date:'medium' }}</td><td><b>{{ event.event_type }}</b></td><td>—</td><td><span class="status">{{ event.status }}</span></td><td>0/5</td><td>Awaiting worker</td></tr> }
                    }
                  </tbody></table></div>
                </section>
              }
            </main>
          </div>
        }
      </div>
    }

    @if (modal) { <div class="modal-backdrop" (click)="closeModal()"><section class="modal" (click)="$event.stopPropagation()"><div class="section-head"><h2>{{ modalTitle }}</h2><button class="icon-button" (click)="closeModal()">×</button></div>
      @if (modal === 'user') { <form (ngSubmit)="saveUser()"><label>Name<input name="name" [(ngModel)]="userForm.name" required /></label><label>Email<input name="email" type="email" [(ngModel)]="userForm.email" required /></label><label>Password <small>{{ userForm.id ? '(leave blank to retain)' : '' }}</small><input name="password" type="password" [(ngModel)]="userForm.password" [required]="!userForm.id" /></label><fieldset><legend>Groups</legend>@for (g of groups; track g.id) { <label class="check"><input type="checkbox" [checked]="userForm.group_ids.includes(g.id)" (change)="toggleGroup(g.id)" />{{ g.name }}</label> }</fieldset><button class="primary">Save user</button></form> }
      @if (modal === 'group') { <form (ngSubmit)="saveGroup()"><label>Name<input name="name" [(ngModel)]="groupForm.name" required /></label><label>Description<textarea name="description" [(ngModel)]="groupForm.description"></textarea></label><button class="primary">Save group</button></form> }
      @if (modal === 'members') { <div class="inline-form"><select [(ngModel)]="selectedUser"><option value="">Select user</option>@for (u of users; track u.id) { <option [value]="u.id">{{ u.name }}</option> }</select><button class="primary small" (click)="addMember()">Add</button></div><ul class="plain-list">@for (u of groupMembers; track u.id) { <li><span>{{ u.name }} <small>{{ u.email }}</small></span><button class="danger" (click)="removeMember(u)">Remove</button></li> }</ul> }
      @if (modal === 'app') { <form (ngSubmit)="saveApp()"><label>Name<input name="name" [(ngModel)]="appForm.name" required /></label><label>Launch URL<input name="launch" type="url" [(ngModel)]="appForm.launch_url" required /></label><label>Logout notification URL<input name="logout" type="url" [(ngModel)]="appForm.logout_notification_url" required /></label>@if (!appForm.id) { <label>Redirect URI<input name="redirect" type="url" [(ngModel)]="appForm.redirect_uri" required /></label> } @else { <label>Status<select name="status" [(ngModel)]="appForm.status"><option>active</option><option>inactive</option></select></label> }<button class="primary">Save application</button></form> }
      @if (modal === 'uris') { <div class="inline-form"><input [(ngModel)]="newURI" type="url" placeholder="https://app.example/callback" /><button class="primary small" (click)="addURI()">Add</button></div><ul class="plain-list">@for (uri of selectedApp.redirect_uri_items || []; track uri.id) { <li><code>{{ uri.redirect_uri }}</code><button class="danger" (click)="removeURI(uri)">Remove</button></li> }</ul> }
    </section></div> }

    @if (showPolicy) { <div class="modal-backdrop"><section class="modal"><div class="section-head"><h2>Add access policy</h2><button class="icon-button" (click)="showPolicy = false">×</button></div><form (ngSubmit)="createPolicy()"><label>Group<select name="group" [(ngModel)]="policyForm.group_id" required><option value="">Choose group</option>@for (g of groups; track g.id) { <option [value]="g.id">{{ g.name }}</option> }</select></label><label>Application<select name="app" [(ngModel)]="policyForm.application_id" required><option value="">Choose application</option>@for (a of apps; track a.id) { <option [value]="a.id">{{ a.name }}</option> }</select></label><button class="primary">Create allow policy</button></form></section></div> }

    @if (createdSecret) { <div class="modal-backdrop"><section class="modal secret-modal"><p class="eyebrow">Shown once</p><h2>Copy the client secret</h2><p>This key cannot be viewed again. Store it securely now.</p><code class="secret">{{ createdSecret }}</code><button class="primary" (click)="copySecret()">{{ copied ? 'Copied' : 'Copy secret' }}</button><button (click)="createdSecret = ''">I have stored it</button></section></div> }
  `,
})
export class PortalComponent implements OnInit {
  private api = inject(ApiService); private router = inject(Router);
  loading = true; profile: any; isAdmin = false; notice = ''; tab: Tab = 'users'; search = ''; statusFilter = '';
  users: any[] = []; groups: any[] = []; apps: any[] = []; policies: any[] = []; audit: any[] = []; events: any[] = [];
  modal = ''; showPolicy = false; selectedGroup: any; selectedApp: any; groupMembers: any[] = []; selectedUser = ''; newURI = '';
  createdSecret = ''; copied = false;
  userForm: any = {}; groupForm: any = {}; appForm: any = {}; policyForm = { group_id: '', application_id: '', effect: 'allow' };
  nav: { id: Tab; number: string; label: string }[] = [{id:'users',number:'I',label:'Users'},{id:'groups',number:'II',label:'Groups'},{id:'apps',number:'III',label:'Applications'},{id:'policies',number:'IV',label:'Policies'},{id:'audit',number:'V',label:'Audit logs'},{id:'events',number:'VI',label:'Deliveries'}];
  get title() { return this.nav.find(n => n.id === this.tab)?.label || ''; }
  get modalTitle() { return this.modal === 'user' ? (this.userForm.id ? 'Edit user' : 'Add user') : this.modal === 'group' ? (this.groupForm.id ? 'Edit group' : 'Add group') : this.modal === 'members' ? `${this.selectedGroup?.name} members` : this.modal === 'app' ? (this.appForm.id ? 'Edit application' : 'Register application') : 'Redirect URIs'; }

  ngOnInit() { this.api.profile().subscribe({ next: r => { this.profile = r.user; this.isAdmin = this.profile.groups?.includes('Admin'); this.loading = false; if (this.isAdmin) this.loadAll(); }, error: () => { this.loading = false; } }); }
  goLogin() { void this.router.navigateByUrl('/login'); }
  logout() { this.api.logout().subscribe({ next: () => this.goLogin(), error: () => this.goLogin() }); }
  select(tab: Tab) { this.tab = tab; this.search = ''; this.load(tab); }
  loadAll() { ['users','groups','apps','policies'].forEach(r => this.load(r as Tab)); }
  load(tab: Tab) { const resource = tab === 'audit' ? 'audit-logs' : tab; this.api.list(resource).subscribe({ next: data => (this as any)[tab] = data || [], error: e => this.fail(e) }); }
  filtered(items: any[]) { const q = this.search.toLowerCase(); return !q ? items : items.filter(item => JSON.stringify(item).toLowerCase().includes(q)); }
  filteredAudit() { return this.filtered(this.audit).filter(l => !this.statusFilter || l.result === this.statusFilter); }
  metadata(value: string) { try { const parsed = JSON.parse(value || '{}'); return parsed.reason || parsed.requestId || '—'; } catch { return '—'; } }
  notify(message: string) { this.notice = message; window.setTimeout(() => this.notice = '', 3000); }
  fail(error: any) { this.notify(errorMessage(error)); if (error.status === 401 || error.status === 403) this.goLogin(); }
  closeModal() { this.modal = ''; }

  openUser(u?: any) { this.userForm = u ? { ...u, password: '', group_ids: (u.groups || []).map((g:any)=>g.id) } : { name:'', email:'', password:'', group_ids:[] }; this.modal = 'user'; }
  toggleGroup(id:string) { const ids=this.userForm.group_ids; this.userForm.group_ids = ids.includes(id) ? ids.filter((x:string)=>x!==id) : [...ids,id]; }
  saveUser() { const {id,...body}=this.userForm; const request=id?this.api.update('users',id,body):this.api.create('users',body); request.subscribe({next:()=>{this.closeModal();this.load('users');this.notify('User saved.');},error:e=>this.fail(e)}); }
  toggleUser(u:any) { this.api.status(u.id,u.status==='active'?'inactive':'active').subscribe({next:()=>{this.load('users');this.notify('User status updated.');},error:e=>this.fail(e)}); }
  deleteUser(u:any) { if(confirm(`Delete ${u.name}?`)) this.api.remove('users',u.id).subscribe({next:()=>{this.load('users');this.notify('User deleted.');},error:e=>this.fail(e)}); }

  openGroup(g?:any) { this.groupForm=g?{...g}:{name:'',description:''};this.modal='group'; }
  saveGroup(){const{id,...body}=this.groupForm;const req=id?this.api.update('groups',id,body):this.api.create('groups',body);req.subscribe({next:()=>{this.closeModal();this.load('groups');this.notify('Group saved.');},error:e=>this.fail(e)});}
  deleteGroup(g:any){if(confirm(`Delete group ${g.name}?`))this.api.remove('groups',g.id).subscribe({next:()=>{this.load('groups');this.notify('Group deleted.');},error:e=>this.fail(e)});}
  manageGroup(g:any){this.selectedGroup=g;this.api.get('groups',g.id).subscribe({next:r=>{this.groupMembers=r.users||[];this.modal='members';},error:e=>this.fail(e)});}
  addMember(){if(!this.selectedUser)return;this.api.assign(this.selectedGroup.id,this.selectedUser).subscribe({next:()=>{this.selectedUser='';this.manageGroup(this.selectedGroup);this.notify('Member added.');},error:e=>this.fail(e)});}
  removeMember(u:any){if(confirm(`Remove ${u.name} from this group?`))this.api.unassign(this.selectedGroup.id,u.id).subscribe({next:()=>{this.manageGroup(this.selectedGroup);this.notify('Member removed.');},error:e=>this.fail(e)});}

  openApp(a?:any){this.appForm=a?{...a}:{name:'',launch_url:'',logout_notification_url:'',redirect_uri:'',status:'active'};this.modal='app';}
  saveApp(){const{id,redirect_uri,redirect_uris,redirect_uri_items,client_id,client_secret_prefix,created_at,...rest}=this.appForm;const body=id?rest:{...rest,redirect_uris:[redirect_uri]};const req=id?this.api.update('apps',id,body):this.api.create('apps',body);req.subscribe({next:(r:any)=>{this.closeModal();this.load('apps');if(r.client_secret)this.createdSecret=r.client_secret;this.notify('Application saved.');},error:e=>this.fail(e)});}
  deleteApp(a:any){if(confirm(`Delete application ${a.name}?`))this.api.remove('apps',a.id).subscribe({next:()=>{this.load('apps');this.notify('Application deleted.');},error:e=>this.fail(e)});}
  manageURIs(a:any){this.api.get('apps',a.id).subscribe({next:r=>{this.selectedApp=r;this.modal='uris';},error:e=>this.fail(e)});}
  addURI(){if(!this.newURI)return;this.api.addURI(this.selectedApp.id,this.newURI).subscribe({next:()=>{this.newURI='';this.manageURIs(this.selectedApp);this.load('apps');this.notify('Redirect URI added.');},error:e=>this.fail(e)});}
  removeURI(uri:any){if(confirm('Remove this redirect URI?'))this.api.removeURI(this.selectedApp.id,uri.id).subscribe({next:()=>{this.manageURIs(this.selectedApp);this.load('apps');this.notify('Redirect URI removed.');},error:e=>this.fail(e)});}
  copySecret(){navigator.clipboard.writeText(this.createdSecret).then(()=>this.copied=true);}

  createPolicy(){this.api.create('policies',this.policyForm).subscribe({next:()=>{this.showPolicy=false;this.load('policies');this.notify('Policy created.');},error:e=>this.fail(e)});}
  deletePolicy(p:any){if(confirm(`Revoke ${p.group_name} access to ${p.application_name}?`))this.api.remove('policies',p.id).subscribe({next:()=>{this.load('policies');this.notify('Policy revoked.');},error:e=>this.fail(e)});}
}
