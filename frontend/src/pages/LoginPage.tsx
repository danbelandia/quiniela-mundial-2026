import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../shared/AuthContext';

export function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const navigate = useNavigate();
  const { login } = useAuth();

  const handleLogin = async () => {
    const res = await fetch('http://localhost:8080/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });

    if (res.ok) {
      const user = await res.json();
      login(String(user.id), !!user.is_admin);
      navigate('/');
    } else {
      alert('Credenciales incorrectas');
    }
  };

  return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <div className="p-6 max-w-sm w-full bg-white rounded shadow-sm border">
        <h2 className="text-xl font-bold mb-4">Iniciar Sesión</h2>
        <input className="w-full border p-2 mb-2" placeholder="Usuario" value={username} onChange={e => setUsername(e.target.value)} />
        <input className="w-full border p-2 mb-4" placeholder="Contraseña" type="password" value={password} onChange={e => setPassword(e.target.value)} />
        <button className="w-full bg-tm-blue text-white p-2 rounded" onClick={handleLogin}>Ingresar</button>
      </div>
    </div>
  );
}
