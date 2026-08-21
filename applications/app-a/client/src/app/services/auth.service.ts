import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private http = inject(HttpClient);
  private backendUrl = 'http://localhost:8081';

  private opts = { withCredentials: true };

  getMe(): Observable<any> {
    return this.http.get(`${this.backendUrl}/me`, this.opts);
  }

  logout(): Observable<any> {
    return this.http.post(`${this.backendUrl}/logout`, {}, this.opts);
  }

  getEvents(): Observable<any> {
    return this.http.get(`${this.backendUrl}/events`, this.opts);
  }

  redirectToLogin() {
    window.location.href = `${this.backendUrl}/auth/login`;
  }

  redirectToSSOLogout() {
    window.location.href = 'http://localhost:8080/logout';
  }
}
