import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private http = inject(HttpClient);
  readonly backend = 'http://localhost:8081';
  readonly sso = 'http://localhost:8080';
  private options = { withCredentials: true };
  me() { return this.http.get<any>(`${this.backend}/me`, this.options); }
  activity() { return this.http.get<any[]>(`${this.backend}/activity`, this.options); }
  events() { return this.http.get<any[]>(`${this.backend}/events`, this.options); }
  status() { return this.http.get<any>(`${this.backend}/session-status`, this.options); }
  localLogout() { return this.http.post(`${this.backend}/logout`, {}, this.options); }
  ssoLogout() { return this.http.post(`${this.sso}/logout`, {}, this.options); }
  login() { window.location.assign(`${this.backend}/auth/login`); }
}

export function errorMessage(error: any) {
  return error?.error?.error?.message || 'The request could not be completed.';
}
