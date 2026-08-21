import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

export interface ApiError { error?: { code?: string; message?: string; requestId?: string } }

@Injectable({ providedIn: 'root' })
export class ApiService {
  private http = inject(HttpClient);
  readonly base = 'http://localhost:8080';
  private options = { withCredentials: true };

  login(body: { email: string; password: string }) { return this.http.post(`${this.base}/login`, body, this.options); }
  logout() { return this.http.post(`${this.base}/logout`, {}, this.options); }
  profile() { return this.http.get<any>(`${this.base}/profile`, this.options); }
  changePassword(body: object) { return this.http.post(`${this.base}/change-password`, body, this.options); }

  list(resource: string) { return this.http.get<any[]>(`${this.base}/admin/${resource}`, this.options); }
  get(resource: string, id: string) { return this.http.get<any>(`${this.base}/admin/${resource}/${id}`, this.options); }
  create(resource: string, body: object) { return this.http.post<any>(`${this.base}/admin/${resource}`, body, this.options); }
  update(resource: string, id: string, body: object) { return this.http.put<any>(`${this.base}/admin/${resource}/${id}`, body, this.options); }
  remove(resource: string, id: string) { return this.http.delete(`${this.base}/admin/${resource}/${id}`, this.options); }
  status(userId: string, status: string) { return this.http.patch(`${this.base}/admin/users/${userId}/status`, { status }, this.options); }
  assign(groupId: string, userId: string) { return this.http.post(`${this.base}/admin/groups/${groupId}/users`, { user_id: userId }, this.options); }
  unassign(groupId: string, userId: string) { return this.http.delete(`${this.base}/admin/groups/${groupId}/users/${userId}`, this.options); }
  addURI(appId: string, uri: string) { return this.http.post<any>(`${this.base}/admin/apps/${appId}/redirect-uris`, { uri }, this.options); }
  removeURI(appId: string, uriId: string) { return this.http.delete(`${this.base}/admin/apps/${appId}/redirect-uris/${uriId}`, this.options); }
}

export function errorMessage(error: ApiError | any): string {
  return error?.error?.error?.message || error?.error?.message || 'The request could not be completed.';
}
