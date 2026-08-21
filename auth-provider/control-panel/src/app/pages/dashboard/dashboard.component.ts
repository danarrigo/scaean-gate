import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="layout">
      <!-- Sidebar -->
      <aside class="sidebar">
        <div class="sidebar-header">
          <span class="gate-icon">🏛️</span>
          <div>
            <h2>Scaean Gate</h2>
            <span class="badge badge-gray">SSO Control Panel</span>
          </div>
        </div>

        <nav class="nav-menu">
          <button [class.active]="activeTab === 'users'" (click)="setTab('users')">👥 Users</button>
          <button [class.active]="activeTab === 'groups'" (click)="setTab('groups')">🏢 Groups</button>
          <button [class.active]="activeTab === 'apps'" (click)="setTab('apps')">🚀 Applications</button>
          <button [class.active]="activeTab === 'policies'" (click)="setTab('policies')">🛡️ Access Policies</button>
          <button [class.active]="activeTab === 'audit'" (click)="setTab('audit')">📜 Audit Logs</button>
          <button [class.active]="activeTab === 'deliveries'" (click)="setTab('deliveries')">📨 Event Deliveries (DLQ)</button>
        </nav>

        <div class="sidebar-footer">
          <button class="btn-logout" (click)="onLogout()">🚪 Sign Out</button>
        </div>
      </aside>

      <!-- Main Content -->
      <main class="main-content">
        <header class="topbar">
          <div class="page-title">
            <h1>{{ getTabTitle() }}</h1>
          </div>
          <div class="user-pill">
            <span class="dot"></span>
            <span>Admin Active</span>
          </div>
        </header>

        @if (notification) {
          <div class="notification-banner">
            <span>{{ notification }}</span>
          </div>
        }

        <div class="content-body">
          <!-- USERS TAB -->
          @if (activeTab === 'users') {
            <div class="section-card">
              <div class="card-header">
                <h3>System Users</h3>
                <button class="btn-sm btn-primary" (click)="showUserModal = true">+ Add User</button>
              </div>

              <table class="data-table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Email</th>
                    <th>Status</th>
                    <th>Created At</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  @for (u of users; track u.id) {
                    <tr>
                      <td class="font-bold">{{ u.name }}</td>
                      <td>{{ u.email }}</td>
                      <td>
                        <span class="badge" [class.badge-active]="u.status === 'active'" [class.badge-danger]="u.status !== 'active'">
                          {{ u.status }}
                        </span>
                      </td>
                      <td class="text-muted">{{ u.created_at | date:'short' }}</td>
                      <td>
                        <button class="btn-xs" (click)="toggleUserStatus(u)">
                          {{ u.status === 'active' ? 'Deactivate' : 'Activate' }}
                        </button>
                      </td>
                    </tr>
                  }
                </tbody>
              </table>
            </div>
          }

          <!-- GROUPS TAB -->
          @if (activeTab === 'groups') {
            <div class="section-card">
              <div class="card-header">
                <h3>User Groups</h3>
                <button class="btn-sm btn-primary" (click)="showGroupModal = true">+ Add Group</button>
              </div>

              <div class="groups-grid">
                @for (g of groups; track g.id) {
                  <div class="group-card">
                    <div class="group-title">
                      <h4>{{ g.name }}</h4>
                    </div>
                    <p class="group-desc">{{ g.description || 'No description provided.' }}</p>
                    <div class="group-actions">
                      <button class="btn-xs btn-outline" (click)="openGroupMembers(g)">Manage Members</button>
                    </div>
                  </div>
                }
              </div>
            </div>
          }

          <!-- APPLICATIONS TAB -->
          @if (activeTab === 'apps') {
            <div class="section-card">
              <div class="card-header">
                <h3>Registered Applications</h3>
              </div>

              <table class="data-table">
                <thead>
                  <tr>
                    <th>App Name</th>
                    <th>Client ID</th>
                    <th>Launch URL</th>
                    <th>Logout Webhook URL</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  @for (a of applications; track a.id) {
                    <tr>
                      <td class="font-bold">{{ a.name }}</td>
                      <td><code>{{ a.client_id }}</code></td>
                      <td><a [href]="a.launch_url" target="_blank">{{ a.launch_url }}</a></td>
                      <td><code>{{ a.logout_notification_url }}</code></td>
                      <td>
                        <span class="badge badge-active">{{ a.status }}</span>
                      </td>
                    </tr>
                  }
                </tbody>
              </table>
            </div>
          }

          <!-- POLICIES TAB -->
          @if (activeTab === 'policies') {
            <div class="section-card">
              <div class="card-header">
                <h3>Access Policy Rules (RBAC)</h3>
                <button class="btn-sm btn-primary" (click)="showPolicyModal = true">+ Add Policy Rule</button>
              </div>

              <table class="data-table">
                <thead>
                  <tr>
                    <th>Group</th>
                    <th>Application</th>
                    <th>Access Effect</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  @for (p of policies; track p.id) {
                    <tr>
                      <td class="font-bold">{{ p.group?.name || p.group_id }}</td>
                      <td>{{ p.application?.name || p.application_id }}</td>
                      <td>
                        <span class="badge" [class.badge-active]="p.effect === 'allow'" [class.badge-danger]="p.effect !== 'allow'">
                          {{ p.effect | uppercase }}
                        </span>
                      </td>
                      <td>
                        <button class="btn-xs btn-danger" (click)="deletePolicy(p.id)">Revoke</button>
                      </td>
                    </tr>
                  }
                </tbody>
              </table>
            </div>
          }

          <!-- AUDIT LOGS TAB -->
          @if (activeTab === 'audit') {
            <div class="section-card">
              <div class="card-header">
                <h3>Security Audit Trail</h3>
                <button class="btn-sm btn-outline" (click)="loadAuditLogs()">🔄 Refresh</button>
              </div>

              <table class="data-table">
                <thead>
                  <tr>
                    <th>Timestamp</th>
                    <th>Event Type</th>
                    <th>Result</th>
                    <th>IP Address</th>
                  </tr>
                </thead>
                <tbody>
                  @for (log of auditLogs; track log.id) {
                    <tr>
                      <td class="text-muted">{{ log.created_at | date:'medium' }}</td>
                      <td class="font-bold">{{ log.event_type }}</td>
                      <td>
                        <span class="badge" [class.badge-active]="log.result === 'success'" [class.badge-danger]="log.result !== 'success'">
                          {{ log.result }}
                        </span>
                      </td>
                      <td><code>{{ log.ip_address || '127.0.0.1' }}</code></td>
                    </tr>
                  }
                </tbody>
              </table>
            </div>
          }

          <!-- DELIVERIES TAB -->
          @if (activeTab === 'deliveries') {
            <div class="section-card">
              <div class="card-header">
                <h3>Sync Worker Delivery Tracker (DLQ)</h3>
                <button class="btn-sm btn-outline" (click)="loadEvents()">🔄 Refresh</button>
              </div>

              <table class="data-table">
                <thead>
                  <tr>
                    <th>Event Type</th>
                    <th>Created</th>
                    <th>Target App</th>
                    <th>Delivery Status</th>
                    <th>Attempts</th>
                    <th>Processed At</th>
                  </tr>
                </thead>
                <tbody>
                  @for (e of events; track e.id) {
                    @if (e.deliveries && e.deliveries.length > 0) {
                      @for (d of e.deliveries; track d.id) {
                        <tr>
                          <td class="font-bold">{{ e.event_type }}</td>
                          <td class="text-muted">{{ e.created_at | date:'short' }}</td>
                          <td><code>{{ d.application?.name || d.application_id }}</code></td>
                          <td>
                            <span class="badge" 
                              [class.badge-succeeded]="d.status === 'succeeded'"
                              [class.badge-retrying]="d.status === 'retrying'"
                              [class.badge-failed]="d.status === 'failed'">
                              {{ d.status | uppercase }}
                            </span>
                          </td>
                          <td><strong>{{ d.attempt_count }} / 5</strong></td>
                          <td class="text-muted">{{ d.processed_at ? (d.processed_at | date:'short') : (d.next_retry_at ? 'Next: ' + (d.next_retry_at | date:'shortTime') : '-') }}</td>
                        </tr>
                      }
                    } @else {
                      <tr>
                        <td class="font-bold">{{ e.event_type }}</td>
                        <td class="text-muted">{{ e.created_at | date:'short' }}</td>
                        <td colspan="4" class="text-muted text-sm">Published to Kafka (Pending Sync Worker pickup)</td>
                      </tr>
                    }
                  }
                </tbody>
              </table>
            </div>
          }
        </div>
      </main>

      <!-- MODALS -->
      @if (showUserModal) {
        <div class="modal-backdrop">
          <div class="modal-box">
            <h3>Add New User</h3>
            <div class="form-group">
              <label>Full Name</label>
              <input type="text" [(ngModel)]="newUser.name" placeholder="John Doe" />
            </div>
            <div class="form-group">
              <label>Email Address</label>
              <input type="email" [(ngModel)]="newUser.email" placeholder="john@example.com" />
            </div>
            <div class="form-group">
              <label>Password</label>
              <input type="password" [(ngModel)]="newUser.password" placeholder="••••••••••••" />
            </div>
            <div class="modal-actions">
              <button class="btn-sm btn-outline" (click)="showUserModal = false">Cancel</button>
              <button class="btn-sm btn-primary" (click)="createUser()">Save User</button>
            </div>
          </div>
        </div>
      }

      @if (showGroupModal) {
        <div class="modal-backdrop">
          <div class="modal-box">
            <h3>Add User Group</h3>
            <div class="form-group">
              <label>Group Name</label>
              <input type="text" [(ngModel)]="newGroup.name" placeholder="Finance Team" />
            </div>
            <div class="form-group">
              <label>Description</label>
              <input type="text" [(ngModel)]="newGroup.description" placeholder="Members of Finance" />
            </div>
            <div class="modal-actions">
              <button class="btn-sm btn-outline" (click)="showGroupModal = false">Cancel</button>
              <button class="btn-sm btn-primary" (click)="createGroup()">Save Group</button>
            </div>
          </div>
        </div>
      }

      @if (showPolicyModal) {
        <div class="modal-backdrop">
          <div class="modal-box">
            <h3>Add Access Policy Rule</h3>
            <div class="form-group">
              <label>Group</label>
              <select [(ngModel)]="newPolicy.groupId">
                @for (g of groups; track g.id) {
                  <option [value]="g.id">{{ g.name }}</option>
                }
              </select>
            </div>
            <div class="form-group">
              <label>Application</label>
              <select [(ngModel)]="newPolicy.applicationId">
                @for (a of applications; track a.id) {
                  <option [value]="a.id">{{ a.name }}</option>
                }
              </select>
            </div>
            <div class="form-group">
              <label>Effect</label>
              <select [(ngModel)]="newPolicy.effect">
                <option value="allow">ALLOW</option>
                <option value="deny">DENY</option>
              </select>
            </div>
            <div class="modal-actions">
              <button class="btn-sm btn-outline" (click)="showPolicyModal = false">Cancel</button>
              <button class="btn-sm btn-primary" (click)="createPolicy()">Save Rule</button>
            </div>
          </div>
        </div>
      }

      @if (selectedGroup) {
        <div class="modal-backdrop">
          <div class="modal-box">
            <h3>Members in {{ selectedGroup.name }}</h3>
            <div class="member-add-row">
              <select [(ngModel)]="selectedUserIdToAdd">
                <option value="">Select User to Add...</option>
                @for (u of users; track u.id) {
                  <option [value]="u.id">{{ u.name }} ({{ u.email }})</option>
                }
              </select>
              <button class="btn-sm btn-primary" (click)="addUserToGroup()">Add</button>
            </div>

            <div class="members-list">
              @for (m of groupMembers; track m.id) {
                <div class="member-item">
                  <span>{{ m.name }} ({{ m.email }})</span>
                  <button class="btn-xs btn-danger" (click)="removeUserFromGroup(m.id)">Remove</button>
                </div>
              }
            </div>

            <div class="modal-actions">
              <button class="btn-sm btn-outline" (click)="selectedGroup = null">Close</button>
            </div>
          </div>
        </div>
      }
    </div>
  `,
  styles: [`
    .layout {
      display: flex;
      min-height: 100vh;
    }

    .sidebar {
      width: 260px;
      background: #1e1b4b;
      color: #f8fafc;
      display: flex;
      flex-direction: column;
      padding: 24px 16px;
      border-right: 1px solid #312e81;
    }

    .sidebar-header {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 32px;
      padding: 0 8px;
    }

    .gate-icon {
      font-size: 2rem;
    }

    .sidebar-header h2 {
      font-size: 1.15rem;
      font-weight: 700;
      color: white;
    }

    .nav-menu {
      display: flex;
      flex-direction: column;
      gap: 6px;
      flex: 1;
    }

    .nav-menu button {
      width: 100%;
      text-align: left;
      padding: 10px 14px;
      background: transparent;
      border: none;
      border-radius: 8px;
      color: #cbd5e1;
      font-size: 0.9rem;
      font-weight: 500;
      transition: all 0.15s ease;
    }

    .nav-menu button:hover {
      background: #312e81;
      color: white;
    }

    .nav-menu button.active {
      background: #4f46e5;
      color: white;
      font-weight: 600;
    }

    .sidebar-footer {
      padding-top: 16px;
      border-top: 1px solid #312e81;
    }

    .btn-logout {
      width: 100%;
      padding: 10px;
      background: #312e81;
      color: #fca5a5;
      border: none;
      border-radius: 8px;
      font-size: 0.85rem;
      font-weight: 600;
    }

    .btn-logout:hover {
      background: #ef4444;
      color: white;
    }

    .main-content {
      flex: 1;
      display: flex;
      flex-direction: column;
      background: #f8fafc;
      overflow-y: auto;
    }

    .topbar {
      height: 70px;
      background: white;
      border-bottom: 1px solid #e2e8f0;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 32px;
    }

    .page-title h1 {
      font-size: 1.35rem;
      font-weight: 700;
      color: #0f172a;
    }

    .user-pill {
      display: flex;
      align-items: center;
      gap: 8px;
      background: #eef2ff;
      color: #4f46e5;
      padding: 6px 14px;
      border-radius: 9999px;
      font-size: 0.85rem;
      font-weight: 600;
    }

    .dot {
      width: 8px;
      height: 8px;
      background: #10b981;
      border-radius: 50%;
    }

    .notification-banner {
      background: #d1fae5;
      color: #065f46;
      padding: 10px 32px;
      font-size: 0.875rem;
      font-weight: 500;
      border-bottom: 1px solid #a7f3d0;
    }

    .content-body {
      padding: 32px;
      display: flex;
      flex-direction: column;
      gap: 24px;
    }

    .section-card {
      background: white;
      border-radius: 12px;
      padding: 24px;
      box-shadow: 0 1px 3px rgba(0,0,0,0.05);
      border: 1px solid #e2e8f0;
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;
    }

    .card-header h3 {
      font-size: 1.1rem;
      font-weight: 700;
      color: #1e293b;
    }

    .data-table {
      width: 100%;
      border-collapse: collapse;
      font-size: 0.9rem;
    }

    .data-table th {
      text-align: left;
      padding: 12px 14px;
      background: #f8fafc;
      color: #64748b;
      font-weight: 600;
      border-bottom: 1px solid #e2e8f0;
    }

    .data-table td {
      padding: 14px;
      border-bottom: 1px solid #f1f5f9;
      color: #334155;
    }

    .font-bold { font-weight: 600; }
    .text-muted { color: #64748b; }
    .text-sm { font-size: 0.8rem; }
    .text-xs { font-size: 0.75rem; }

    .groups-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
      gap: 16px;
    }

    .group-card {
      border: 1px solid #e2e8f0;
      border-radius: 10px;
      padding: 18px;
      background: #fafafa;
    }

    .group-title h4 {
      font-size: 1rem;
      font-weight: 700;
      color: #0f172a;
    }

    .group-desc {
      font-size: 0.85rem;
      color: #64748b;
      margin: 8px 0 14px;
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

    .btn-primary { background: #4f46e5; color: white; }
    .btn-primary:hover { background: #4338ca; }
    .btn-danger { background: #ef4444; color: white; border: none; }
    .btn-danger:hover { background: #dc2626; }
    .btn-outline { background: white; border: 1px solid #cbd5e1; color: #334155; }
    .btn-outline:hover { background: #f8fafc; }

    .modal-backdrop {
      position: fixed;
      inset: 0;
      background: rgba(0, 0, 0, 0.5);
      display: flex;
      align-items: center;
      justify-content: center;
      z-index: 50;
    }

    .modal-box {
      background: white;
      width: 100%;
      max-width: 460px;
      border-radius: 12px;
      padding: 24px;
      box-shadow: 0 20px 25px -5px rgba(0,0,0,0.2);
    }

    .modal-box h3 {
      font-size: 1.15rem;
      font-weight: 700;
      margin-bottom: 16px;
    }

    .form-group {
      margin-bottom: 16px;
    }

    .form-group label {
      display: block;
      font-size: 0.85rem;
      font-weight: 500;
      color: #475569;
      margin-bottom: 6px;
    }

    .form-group input, .form-group select {
      width: 100%;
      padding: 9px 12px;
      border: 1.5px solid #cbd5e1;
      border-radius: 6px;
      font-size: 0.9rem;
      outline: none;
    }

    .modal-actions {
      display: flex;
      justify-content: flex-end;
      gap: 10px;
      margin-top: 20px;
    }

    .member-add-row {
      display: flex;
      gap: 8px;
      margin-bottom: 16px;
    }

    .member-add-row select {
      flex: 1;
      padding: 8px 10px;
      border-radius: 6px;
      border: 1px solid #cbd5e1;
    }

    .members-list {
      max-height: 200px;
      overflow-y: auto;
      border: 1px solid #e2e8f0;
      border-radius: 6px;
      padding: 8px;
    }

    .member-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 6px 8px;
      border-bottom: 1px solid #f1f5f9;
      font-size: 0.85rem;
    }
  `]
})
export class DashboardComponent implements OnInit {
  private api = inject(ApiService);
  private router = inject(Router);

  activeTab = 'users';
  notification = '';

  users: any[] = [];
  groups: any[] = [];
  applications: any[] = [];
  policies: any[] = [];
  auditLogs: any[] = [];
  events: any[] = [];

  showUserModal = false;
  newUser = { name: '', email: '', password: '' };

  showGroupModal = false;
  newGroup = { name: '', description: '' };

  showPolicyModal = false;
  newPolicy = { groupId: '', applicationId: '', effect: 'allow' };

  selectedGroup: any = null;
  groupMembers: any[] = [];
  selectedUserIdToAdd = '';

  ngOnInit() {
    this.loadUsers();
    this.loadGroups();
    this.loadApps();
    this.loadPolicies();
  }

  setTab(tab: string) {
    this.activeTab = tab;
    if (tab === 'users') this.loadUsers();
    if (tab === 'groups') this.loadGroups();
    if (tab === 'apps') this.loadApps();
    if (tab === 'policies') this.loadPolicies();
    if (tab === 'audit') this.loadAuditLogs();
    if (tab === 'deliveries') this.loadEvents();
  }

  getTabTitle() {
    switch (this.activeTab) {
      case 'users': return 'User Management';
      case 'groups': return 'Group & Team Management';
      case 'apps': return 'Registered Applications';
      case 'policies': return 'RBAC Access Policies';
      case 'audit': return 'Security Audit Trail';
      case 'deliveries': return 'Sync Worker Event Deliveries (DLQ)';
      default: return 'Control Panel';
    }
  }

  showNotification(msg: string) {
    this.notification = msg;
    setTimeout(() => this.notification = '', 4000);
  }

  loadUsers() {
    this.api.getUsers().subscribe({
      next: (res) => this.users = Array.isArray(res) ? res : (res?.data || []),
      error: (err) => {
        if (err.status === 401 || err.status === 403) {
          this.router.navigate(['/login']);
        }
      }
    });
  }

  toggleUserStatus(user: any) {
    const nextStatus = user.status === 'active' ? 'deactivated' : 'active';
    this.api.updateUserStatus(user.id, nextStatus).subscribe({
      next: () => {
        user.status = nextStatus;
        this.showNotification(`User status updated to ${nextStatus}!`);
      },
      error: (err) => alert(err.error?.error?.message || 'Failed to update user status')
    });
  }

  createUser() {
    this.api.createUser(this.newUser).subscribe({
      next: () => {
        this.showUserModal = false;
        this.newUser = { name: '', email: '', password: '' };
        this.loadUsers();
        this.showNotification('User created successfully!');
      },
      error: (err) => alert(err.error?.error?.message || 'Failed to create user')
    });
  }

  loadGroups() {
    this.api.getGroups().subscribe({
      next: (res) => this.groups = Array.isArray(res) ? res : (res?.data || []),
      error: (err) => console.error('Failed to load groups:', err)
    });
  }

  createGroup() {
    this.api.createGroup(this.newGroup).subscribe({
      next: () => {
        this.showGroupModal = false;
        this.newGroup = { name: '', description: '' };
        this.loadGroups();
        this.showNotification('Group created successfully!');
      },
      error: (err) => alert(err.error?.error?.message || 'Failed to create group')
    });
  }

  openGroupMembers(group: any) {
    this.selectedGroup = group;
    this.api.getGroup(group.id).subscribe({
      next: (res) => this.groupMembers = res?.users || res?.members || [],
      error: (err) => console.error('Failed to load group:', err)
    });
  }

  addUserToGroup() {
    if (!this.selectedUserIdToAdd) return;
    this.api.addGroupMember(this.selectedGroup.id, this.selectedUserIdToAdd).subscribe({
      next: () => {
        this.openGroupMembers(this.selectedGroup);
        this.selectedUserIdToAdd = '';
        this.showNotification('Member added to group!');
      },
      error: (err) => alert(err.error?.error?.message || 'Failed to add member to group')
    });
  }

  removeUserFromGroup(userId: string) {
    this.api.removeGroupMember(this.selectedGroup.id, userId).subscribe({
      next: () => {
        this.openGroupMembers(this.selectedGroup);
        this.showNotification('Member removed from group!');
      },
      error: (err) => alert(err.error?.error?.message || 'Failed to remove member')
    });
  }

  loadApps() {
    this.api.getApps().subscribe({
      next: (res) => this.applications = Array.isArray(res) ? res : (res?.data || []),
      error: (err) => console.error('Failed to load apps:', err)
    });
  }

  loadPolicies() {
    this.api.getPolicies().subscribe({
      next: (res) => this.policies = Array.isArray(res) ? res : (res?.data || []),
      error: (err) => console.error('Failed to load policies:', err)
    });
  }

  createPolicy() {
    this.api.createPolicy(this.newPolicy).subscribe({
      next: () => {
        this.showPolicyModal = false;
        this.loadPolicies();
        this.showNotification('Policy rule created!');
      },
      error: (err) => alert(err.error?.error?.message || 'Failed to create policy rule')
    });
  }

  deletePolicy(policyId: string) {
    this.api.deletePolicy(policyId).subscribe({
      next: () => {
        this.loadPolicies();
        this.showNotification('Policy rule revoked!');
      },
      error: (err) => alert(err.error?.error?.message || 'Failed to delete policy rule')
    });
  }

  loadAuditLogs() {
    this.api.getAuditLogs().subscribe({
      next: (res) => this.auditLogs = Array.isArray(res) ? res : (res?.data || []),
      error: (err) => console.error('Failed to load audit logs:', err)
    });
  }

  loadEvents() {
    this.api.getEvents().subscribe({
      next: (res) => this.events = Array.isArray(res) ? res : (res?.data || []),
      error: (err) => console.error('Failed to load events:', err)
    });
  }

  onLogout() {
    this.api.logout().subscribe(() => {
      this.router.navigate(['/login']);
    });
  }
}
