// web/app/src/routes/router.tsx — complete route table (RFC-010 K.2, ECD-009).
// TanStack Router; lazy-loaded feature bundles.
import { createRootRoute, createRoute, createRouter, Outlet, lazyRouteComponent } from '@tanstack/react-router';
import { AppShell } from '../components/AppShell';

const root = createRootRoute({ component: () => (<AppShell><Outlet /></AppShell>) });

const r = (path: string, feature: string) =>
  createRoute({ getParentRoute: () => root, path,
    component: lazyRouteComponent(() => import(`../features/${feature}/Page`)) });

export const routes = [
  r('/', 'dashboard'),
  r('/kernels', 'kernels'),
  r('/kernels/$hash', 'kernels'),
  r('/compilers', 'compilers'),
  r('/compilers/regressions', 'compilers'),
  r('/graph', 'graph'),
  r('/cost', 'cost'),
  r('/savings', 'savings'),
  r('/simulate', 'simulate'),
  r('/recommendations', 'recs'),
  r('/recommendations/$id', 'recs'),
  r('/governance', 'governance'),
  r('/governance/audit', 'governance'),
  r('/agents', 'agents'),
  r('/agents/tasks/$id', 'agents'),
  r('/reports', 'reports'),
  r('/admin', 'admin'),
  r('/admin/users', 'admin'),
  r('/admin/apikeys', 'admin'),
  r('/admin/clusters', 'admin'),
  r('/notifications', 'notifications'),
];
export const router = createRouter({ routeTree: root.addChildren(routes) });
