import { useState, useEffect, useMemo } from 'react';
import { useMatches } from '../features/hooks';

export function AdminPage() {
  const { matches, loading, refresh } = useMatches();
  const [scores, setScores] = useState<{ [key: number]: { home: number; away: number } }>({});
  const [users, setUsers] = useState<any[]>([]);
  const [editingUserId, setEditingUserId] = useState<number | null>(null);
  const [editUserForm, setEditUserForm] = useState<{ username: string, email: string, is_admin: boolean }>({ username: '', email: '', is_admin: false });
  const [currentGroup, setCurrentGroup] = useState('A');

  useEffect(() => {
    fetch('http://localhost:8080/ranking').then(res => res.json()).then(setUsers);
  }, []);

  const groups = useMemo(
    () => Array.from(new Set(matches.map((m: any) => m.group))).sort(),
    [matches]
  );

  const paginatedMatches = useMemo(
    () => matches.filter((m: any) => m.group === currentGroup),
    [matches, currentGroup]
  );

  useEffect(() => {
    if (groups.length > 0 && !groups.includes(currentGroup)) {
      setCurrentGroup(groups[0]);
    }
  }, [groups, currentGroup]);

  const currentIndex = groups.indexOf(currentGroup);

  const handleUpdate = async (matchId: number) => {
    const s = scores[matchId] || { home: 0, away: 0 };
    try {
      const response = await fetch('http://localhost:8080/admin/matches/result', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: matchId, home_score: s.home, away_score: s.away })
      });
      if (response.ok) {
        alert('Resultado oficial guardado');
        refresh();
      }
    } catch (error) {
      alert('Error al guardar');
    }
  };

  const handleDeleteUser = async (id: number) => {
    if(confirm('¿Eliminar usuario?')) {
      await fetch(`http://localhost:8080/admin/users/${id}`, { method: 'DELETE' });
      setUsers(users.filter(u => u.id !== id));
    }
  };

  const startEdit = (u: any) => {
    setEditingUserId(u.id);
    setEditUserForm({ username: u.username, email: u.email, is_admin: u.is_admin });
  };

  const saveEdit = async (id: number) => {
    try {
      const response = await fetch('http://localhost:8080/admin/users/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id, ...editUserForm })
      });
      if (response.ok) {
        alert('Usuario actualizado');
        setUsers(prev => prev.map(usr => usr.id === id ? { ...usr, ...editUserForm } : usr));
        setEditingUserId(null);
      }
    } catch (error) {
      alert('Error al actualizar usuario');
    }
  };

  if (loading) return <div>Cargando...</div>;

  return (
    <div className="bg-white p-6 rounded-lg shadow-sm border">
      <div className="flex justify-between items-center mb-6 border-b-2 border-tm-blue pb-4">
        <h1 className="text-2xl font-bold text-tm-blue">Panel de Administrador</h1>
        <p className="text-xs text-gray-500 italic">
          El bloqueo de pronósticos se gestiona desde el Cronograma.
        </p>
      </div>

      <h2 className="text-xl font-bold mb-4 text-tm-blue">Cargar Resultados Oficiales</h2>
      <table className="w-full text-left mb-4">
        <thead>
          <tr className="bg-gray-100 text-tm-blue uppercase text-sm">
            <th className="p-3">Partido</th>
            <th className="p-3">Resultado</th>
            <th className="p-3">Acción</th>
          </tr>
        </thead>
        <tbody>
          {paginatedMatches.map((m: any) => (
            <tr key={m.id} className="border-b">
              <td className="p-3">{m.home_team} vs {m.away_team}</td>
              <td className="p-3 flex gap-2">
                <input type="number" className="w-12 border p-1" defaultValue={m.home_score} onChange={e => setScores({...scores, [m.id]: {...scores[m.id], home: parseInt(e.target.value) || 0}})} />
                <input type="number" className="w-12 border p-1" defaultValue={m.away_score} onChange={e => setScores({...scores, [m.id]: {...scores[m.id], away: parseInt(e.target.value) || 0}})} />
              </td>
              <td className="p-3 flex gap-2">
                <button onClick={() => handleUpdate(m.id)} className="bg-tm-blue text-white px-3 py-1 rounded">Actualizar</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      <div className="flex justify-between items-center mb-10">
        <span className="text-sm text-gray-600">
          Grupo {currentGroup} ({paginatedMatches.length} partidos)
        </span>
        <div className="flex gap-2">
          <button
            onClick={() => setCurrentGroup(groups[Math.max(0, currentIndex - 1)])}
            disabled={currentIndex <= 0}
            className="px-3 py-1 rounded border bg-gray-100 hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            ← Anterior
          </button>
          {groups.map(g => (
            <button
              key={g}
              onClick={() => setCurrentGroup(g)}
              className={`px-3 py-1 rounded border font-bold ${currentGroup === g ? 'bg-tm-blue text-white' : 'bg-gray-100 hover:bg-gray-200'}`}
            >
              {g}
            </button>
          ))}
          <button
            onClick={() => setCurrentGroup(groups[Math.min(groups.length - 1, currentIndex + 1)])}
            disabled={currentIndex >= groups.length - 1}
            className="px-3 py-1 rounded border bg-gray-100 hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Siguiente →
          </button>
        </div>
      </div>

      <h2 className="text-xl font-bold mb-4 text-tm-blue">Gestionar Usuarios</h2>
      <table className="w-full text-left">
        <thead>
          <tr className="bg-gray-100 text-tm-blue uppercase text-sm">
            <th className="p-3">Usuario</th>
            <th className="p-3">Email</th>
            <th className="p-3">Acción</th>
          </tr>
        </thead>
        <tbody>
          {users.map(u => (
            <tr key={u.id} className="border-b">
              <td className="p-3">
                {editingUserId === u.id ?
                  <input className="border p-1" value={editUserForm.username} onChange={e => setEditUserForm({...editUserForm, username: e.target.value})} />
                  : u.username
                }
              </td>
              <td className="p-3">
                {editingUserId === u.id ?
                  <input className="border p-1" value={editUserForm.email} onChange={e => setEditUserForm({...editUserForm, email: e.target.value})} />
                  : u.email
                }
              </td>
              <td className="p-3">
                  <input type="checkbox" checked={editingUserId === u.id ? editUserForm.is_admin : u.is_admin} onChange={async (e) => {
                      const newIsAdmin = e.target.checked;
                      if (editingUserId === u.id) {
                          setEditUserForm(prev => ({ ...prev, is_admin: newIsAdmin }));
                      } else {
                          setUsers(prev => prev.map(usr => usr.id === u.id ? { ...usr, is_admin: newIsAdmin } : usr));
                          try {
                              const response = await fetch('http://localhost:8080/admin/users/update', {
                                  method: 'POST',
                                  headers: {'Content-Type': 'application/json'},
                                  body: JSON.stringify({ ...u, is_admin: newIsAdmin })
                              });
                              if (!response.ok) {
                                  setUsers(prev => prev.map(usr => usr.id === u.id ? { ...usr, is_admin: u.is_admin } : usr));
                                  alert('Error al actualizar el rol de admin');
                              }
                          } catch (error) {
                              setUsers(prev => prev.map(usr => usr.id === u.id ? { ...usr, is_admin: u.is_admin } : usr));
                              alert('Error de red al actualizar admin');
                          }
                      }
                  }} />
              </td>
              <td className="p-3 flex gap-2">
                {editingUserId === u.id ?
                  <button onClick={() => saveEdit(u.id)} className="bg-green-600 text-white px-2 py-1 rounded">Guardar</button>
                  : <button onClick={() => startEdit(u)} className="bg-yellow-500 text-white px-2 py-1 rounded">Editar</button>
                }
                <button onClick={() => handleDeleteUser(u.id)} className="bg-red-500 text-white px-2 py-1 rounded">Eliminar</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
