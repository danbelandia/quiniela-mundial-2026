import { Navigate, Outlet } from 'react-router-dom';

export function ProtectedRoute() {
  const userId = localStorage.getItem('userId');
  return userId ? <Outlet /> : <Navigate to="/login" />;
}
