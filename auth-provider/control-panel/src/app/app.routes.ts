import { Routes } from '@angular/router';
import { LoginComponent } from './pages/login.component';
import { PortalComponent } from './pages/portal.component';

export const routes: Routes = [
  { path: 'login', component: LoginComponent },
  { path: '', component: PortalComponent },
  { path: '**', redirectTo: '' },
];
