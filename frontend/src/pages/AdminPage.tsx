import { useState, useEffect, useMemo } from 'react';
import { useMatches, useConfig } from '../features/hooks';
import { apiClient } from '../shared/api/apiClient';

export function AdminPage() {
  const { matches, loading, refresh } = useMatches();
  const { config } = useConfig();
  const [scores, setScores] = useState<{ [key: number]: { home: number; away: number } }>({});
  const [users, setUsers] = useState<any[]>([]);
  const [editingUserId, setEditingUserId] = useState<number | null>(null);
  const [editUserForm, setEditUserForm] = useState<{ username: string, email: string, is_admin: boolean }>({ username: '', email: '', is_admin: false });
  const [currentGroup, setCurrentGroup] = useState('A');
  const [topScorerPick, setTopScorerPick] = useState('');
  const [topScorerCustom, setTopScorerCustom] = useState('');

  useEffect(() => {
    apiClient.get('/ranking').then(setUsers);
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
      await apiClient.post('/admin/matches/result', { id: matchId, home_score: s.home, away_score: s.away });
      alert('Resultado oficial guardado');
      refresh();
    } catch (error) {
      alert('Error al guardar');
    }
  };

  const handleDeleteUser = async (id: number) => {
    if(confirm('¿Eliminar usuario?')) {
      try {
        await apiClient.delete(`/admin/users/${id}`);
        setUsers(users.filter(u => u.id !== id));
      } catch (error) {
        alert('Error al eliminar usuario');
      }
    }
  };

  const startEdit = (u: any) => {
    setEditingUserId(u.id);
    setEditUserForm({ username: u.username, email: u.email, is_admin: u.is_admin });
  };

  const saveEdit = async (id: number) => {
    try {
      await apiClient.post('/admin/users/update', { id, ...editUserForm });
      alert('Usuario actualizado');
      setUsers(prev => prev.map(usr => usr.id === id ? { ...usr, ...editUserForm } : usr));
      setEditingUserId(null);
    } catch (error) {
      alert('Error al actualizar usuario');
    }
  };

  const handleSetTopScorer = async () => {
    const player = topScorerCustom.trim() || topScorerPick;
    if (!player) {
      alert('Elegí un jugador de la lista o escribí uno personalizado');
      return;
    }
    if (topScorerCustom.trim()) {
      if (!confirm(`"${player}" no está en la lista de candidatos. ¿Guardarlo igual?`)) return;
    }
    try {
      await apiClient.post('/admin/top-scorer', { player });
      alert('Goleador guardado');
      setTopScorerCustom('');
    } catch (error) {
      alert('Error al guardar');
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

      <h2 className="text-xl font-bold mb-4 text-tm-blue mt-10">Goleador de la fase de grupos</h2>
      <p className="text-sm text-gray-600 mb-3">
        Al finalizar la fase de grupos, elegí el jugador que más goles haya hecho. Quien acertó suma 6 puntos.
      </p>
      <div className="flex flex-col md:flex-row gap-3 mb-6 items-start md:items-center">
        <select
          value={topScorerPick}
          onChange={e => setTopScorerPick(e.target.value)}
          className="border border-gray-300 rounded p-2 md:w-96"
        >
          <option value="">— De la lista de candidatos —</option>
          {(config?.top_scorer_candidates || []).map((c: any) => (
            <option key={c.name} value={c.name}>{c.flag} {c.name} ({c.team})</option>
          ))}
        </select>
        <span className="text-sm text-gray-500">o</span>
        <input
          type="text"
          placeholder="Nombre personalizado (no está en la lista)"
          value={topScorerCustom}
          onChange={e => setTopScorerCustom(e.target.value)}
          className="border border-gray-300 rounded p-2 md:w-96"
        />
        <button onClick={handleSetTopScorer} className="bg-tm-blue text-white px-4 py-2 rounded font-bold hover:bg-blue-900">
          Guardar resultado
        </button>
      </div>

      <h2 className="text-xl font-bold mb-4 text-tm-blue mt-10">Gestionar Usuarios</h2>
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
                              await apiClient.post('/admin/users/update', { ...u, is_admin: newIsAdmin });
                          } catch (error) {
                              setUsers(prev => prev.map(usr => usr.id === u.id ? { ...usr, is_admin: u.is_admin } : usr));
                              alert('Error al actualizar el rol de admin');
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
