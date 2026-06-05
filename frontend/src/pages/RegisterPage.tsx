import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '../shared/api/apiClient';

export function RegisterPage() {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const navigate = useNavigate();

  const handleRegister = async () => {
    try {
      await apiClient.post('/register', { username, email, password });
      navigate('/ranking');
    } catch {
      alert('Error en el registro');
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh]">
      <div className="p-6 max-w-sm w-full bg-white rounded shadow-sm border">
        <h2 className="text-xl font-bold mb-4">Registro</h2>
        <input className="w-full border p-2 mb-2" placeholder="Usuario" value={username} onChange={e => setUsername(e.target.value)} />
        <input className="w-full border p-2 mb-2" placeholder="Email" value={email} onChange={e => setEmail(e.target.value)} />
        <input className="w-full border p-2 mb-4" placeholder="Password" type="password" value={password} onChange={e => setPassword(e.target.value)} />
        <button className="w-full bg-tm-blue text-white p-2 rounded" onClick={handleRegister}>Registrarse</button>
      </div>

      <div className="max-w-sm w-full mt-6 text-center text-sm text-gray-700">
        <h3 className="font-bold text-tm-blue mb-2">¿Qué es Quiniela Mundial 2026?</h3>
        <p>
          Es una plataforma para grupos cerrados de amigos que compiten pronosticando los resultados de la Copa Mundial de la FIFA 2026. Registrate, ingresá tus pronósticos antes de la fase de grupos y subí en el ranking según los puntos que aciertes.
        </p>
      </div>
    </div>
  );
}