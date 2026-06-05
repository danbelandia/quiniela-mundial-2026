import { BrowserRouter, Routes, Route, Link, useNavigate } from 'react-router-dom';
import { HomePage } from '../pages/HomePage';
import { RankingPage } from '../pages/RankingPage';
import { RegisterPage } from '../pages/RegisterPage';
import { AdminPage } from '../pages/AdminPage';
import { LoginPage } from '../pages/LoginPage';
import { ProtectedRoute } from '../shared/ProtectedRoute';
import { AdminRoute } from '../shared/AdminRoute';
import { AuthProvider, useAuth } from '../shared/AuthContext';

function Navbar() {
  const { isLoggedIn, isAdmin, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <nav className="bg-tm-blue text-white p-4 shadow-md">
      <div className="max-w-4xl mx-auto flex justify-between items-center">
        <div className="flex gap-4">
          <Link to="/" className="font-bold hover:text-gray-300">Cronograma</Link>
          <Link to="/ranking" className="font-bold hover:text-gray-300">Ranking</Link>
          {isLoggedIn ? (
              <>
                  {isAdmin && <Link to="/admin" className="font-bold hover:text-gray-300">Admin</Link>}
              </>
          ) : (
              <>
                  <Link to="/login" className="font-bold hover:text-gray-300">Login</Link>
                  <Link to="/register" className="font-bold hover:text-gray-300">Registro</Link>
              </>
          )}
        </div>
        {isLoggedIn && (
          <button onClick={handleLogout} className="font-bold hover:text-gray-300">Cerrar Sesión</button>
        )}
      </div>
    </nav>
  );
}

export function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <div className="min-h-screen flex flex-col bg-tm-light-blue text-tm-blue">
          <Navbar />
          <main className="max-w-4xl mx-auto p-4 flex-1 w-full">
            <Routes>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />

              <Route element={<ProtectedRoute />}>
                  <Route path="/" element={<HomePage />} />
                  <Route path="/ranking" element={<RankingPage />} />

                  <Route element={<AdminRoute />}>
                      <Route path="/admin" element={<AdminPage />} />
                  </Route>
              </Route>
            </Routes>
          </main>
          <footer className="bg-tm-blue text-white text-center py-4 mt-8">
            <p className="text-sm">Página web creada por <span className="font-bold">Dangel Belandia &amp; CompañIA</span></p>
          </footer>
        </div>
      </BrowserRouter>
    </AuthProvider>
  );
}
