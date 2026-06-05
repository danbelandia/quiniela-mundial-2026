import { Navigate, Outlet } from 'react-router-dom';

export function AdminRoute() {
  const isAdmin = localStorage.getItem('isAdmin') === 'true';
  return isAdmin ? <Outlet /> : <Navigate to="/" />;
}
