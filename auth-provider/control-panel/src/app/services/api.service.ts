import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  private http = inject(HttpClient);
  private baseUrl = 'http://localhost:8080';

  private opts = { withCredentials: true };

  login(body: { email: string; password: string }): Observable<any> {
    return this.http.post(`${this.baseUrl}/login`, body, this.opts);
  }

  logout(): Observable<any> {
    return this.http.post(`${this.baseUrl}/logout`, {}, this.opts);
  }

  getProfile(): Observable<any> {
    return this.http.get(`${this.baseUrl}/profile`, this.opts);
  }

  changePassword(body: { currentPassword: string; newPassword: string }): Observable<any> {
    return this.http.post(`${this.baseUrl}/change-password`, body, this.opts);
  }

  getUsers(): Observable<any> {
    return this.http.get(`${this.baseUrl}/admin/users`, this.opts);
  }

  createUser(body: { name: string; email: string; password: string }): Observable<any> {
    return this.http.post(`${this.baseUrl}/admin/users`, body, this.opts);
  }

  updateUserStatus(id: string, status: string): Observable<any> {
    return this.http.put(`${this.baseUrl}/admin/users/${id}/status`, { status }, this.opts);
  }

  getGroups(): Observable<any> {
    return this.http.get(`${this.baseUrl}/admin/groups`, this.opts);
  }

  createGroup(body: { name: string; description: string }): Observable<any> {
    return this.http.post(`${this.baseUrl}/admin/groups`, body, this.opts);
  }

  getGroupMembers(groupId: string): Observable<any> {
    return this.http.get(`${this.baseUrl}/admin/groups/${groupId}/members`, this.opts);
  }

  addGroupMember(groupId: string, userId: string): Observable<any> {
    return this.http.post(`${this.baseUrl}/admin/groups/${groupId}/members`, { userId }, this.opts);
  }

  removeGroupMember(groupId: string, userId: string): Observable<any> {
    return this.http.delete(`${this.baseUrl}/admin/groups/${groupId}/members?userId=${userId}`, this.opts);
  }

  getApps(): Observable<any> {
    return this.http.get(`${this.baseUrl}/admin/applications`, this.opts);
  }

  createApp(body: any): Observable<any> {
    return this.http.post(`${this.baseUrl}/admin/applications`, body, this.opts);
  }

  getPolicies(): Observable<any> {
    return this.http.get(`${this.baseUrl}/admin/policies`, this.opts);
  }

  createPolicy(body: { groupId: string; applicationId: string; effect: string }): Observable<any> {
    return this.http.post(`${this.baseUrl}/admin/policies`, body, this.opts);
  }

  deletePolicy(policyId: string): Observable<any> {
    return this.http.delete(`${this.baseUrl}/admin/policies/${policyId}`, this.opts);
  }

  getAuditLogs(): Observable<any> {
    return this.http.get(`${this.baseUrl}/admin/audit-logs`, this.opts);
  }

  getEvents(): Observable<any> {
    return this.http.get(`${this.baseUrl}/admin/events`, this.opts);
  }

  getSessions(): Observable<any> {
    return this.http.get(`${this.baseUrl}/admin/sessions`, this.opts);
  }

  revokeSession(sessionId: string): Observable<any> {
    return this.http.post(`${this.baseUrl}/admin/sessions/${sessionId}/revoke`, {}, this.opts);
  }
}
