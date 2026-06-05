import { useState, useMemo, useEffect } from 'react';
import { useMatches, useSubmitPrediction } from '../features/hooks';
import { useAuth } from '../shared/AuthContext';

export function HomePage() {
  const { matches, loading, refresh } = useMatches();
  const { submit } = useSubmitPrediction();
  const { isAdmin } = useAuth();
  const [currentGroup, setCurrentGroup] = useState('A');
  const [editingId, setEditingId] = useState<number | null>(null);
  const [scores, setScores] = useState<{ [key: number]: { home: number; away: number } }>({});
  const [savedScores, setSavedScores] = useState<{ [key: number]: { home: number; away: number } }>({});

  useEffect(() => {
    fetch('http://localhost:8080/predictions')
      .then(res => res.json())
      .then(data => {
        if (!data) return;
        const userId = parseInt(localStorage.getItem('userId') || '1');
        const map = data
          .filter((p: any) => p.user_id === userId)
          .reduce((acc: any, p: any) => {
            acc[p.match_id] = { home: p.home_score, away: p.away_score };
            return acc;
          }, {});
        setSavedScores(map);
      });
  }, []);

  const groups = useMemo(() => {
    const uniqueGroups = Array.from(new Set(matches.map((m: any) => m.group))).sort();
    return uniqueGroups;
  }, [matches]);

  const filteredMatches = useMemo(() => {
    return matches.filter((m: any) => m.group === currentGroup);
  }, [matches, currentGroup]);

  const handleSave = async (matchId: number) => {
    const s = scores[matchId] || { home: 0, away: 0 };
    const userId = localStorage.getItem('userId') || '1';

    try {
      const response = await fetch('http://localhost:8080/predictions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          match_id: matchId,
          home_score: s.home,
          away_score: s.away,
          user_id: parseInt(userId)
        })
      });

      if (!response.ok) throw new Error(`Error ${response.status}`);

      setSavedScores(prev => ({...prev, [matchId]: s}));
      setEditingId(null);
    } catch (error) {
      console.error('Fallo al guardar:', error);
      alert('Error al guardar la predicción');
    }
  };

  const handleLockAll = async (locked: boolean) => {
    try {
      const response = await fetch('http://localhost:8080/admin/matches/lock-all', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ is_locked: locked })
      });
      if (response.ok) {
        alert('Todos los partidos han sido ' + (locked ? 'bloqueados' : 'desbloqueados'));
        refresh();
      }
    } catch (error) {
      alert('Error al actualizar');
    }
  };

  if (loading) return <div className="text-center mt-10 text-tm-blue">Cargando cronograma...</div>;

  return (
    <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-100">
      <div className="flex justify-between items-center border-b-2 border-tm-blue pb-4 mb-6">
        <h1 className="text-2xl font-bold text-tm-blue">Grupo {currentGroup}</h1>
        <div className="flex gap-2">
            <button
                onClick={() => setCurrentGroup(groups[Math.max(0, groups.indexOf(currentGroup) - 1)])}
                className="bg-gray-100 px-3 py-1 rounded hover:bg-gray-200 text-tm-blue font-medium"
            >
                Anterior
            </button>
            <button
                onClick={() => setCurrentGroup(groups[Math.min(groups.length - 1, groups.indexOf(currentGroup) + 1)])}
                className="bg-tm-blue text-white px-3 py-1 rounded hover:bg-blue-900 font-medium"
            >
                Siguiente
            </button>
        </div>
      </div>

      {isAdmin && (
        <div className="flex gap-2 mb-4 p-3 bg-gray-50 rounded border">
          <span className="text-sm text-gray-700 font-medium self-center">Solo admin:</span>
          <button
            onClick={() => handleLockAll(true)}
            className="bg-red-600 text-white px-3 py-1 rounded font-bold text-sm hover:bg-red-700"
          >
            Bloquear Todos
          </button>
          <button
            onClick={() => handleLockAll(false)}
            className="bg-green-600 text-white px-3 py-1 rounded font-bold text-sm hover:bg-green-700"
          >
            Desbloquear Todos
          </button>
        </div>
      )}
      
      <table className="w-full text-left">
        <thead>
          <tr className="bg-gray-100 text-tm-blue uppercase text-xs tracking-wider">
            <th className="p-3">Partido</th>
            <th className="p-3">Pronóstico</th>
          </tr>
        </thead>
        <tbody>
          {filteredMatches.map((m: any) => (
            <tr key={m.id} className="border-b hover:bg-tm-light-blue transition-colors">
              <td className="p-3 font-medium text-gray-800">
                {m.home_flag} {m.home_team} vs {m.away_team} {m.away_flag}
              </td>
              <td className="p-3">
                {editingId === m.id ? (
                  <div className="flex gap-2 items-center">
                    <input type="number" min="0" className="w-12 border rounded p-1 text-center" 
                           defaultValue={scores[m.id]?.home || 0}
                           onChange={e => setScores(prev => ({...prev, [m.id]: {...prev[m.id], home: parseInt(e.target.value) || 0}}))} />
                    <span className="text-gray-400">-</span>
                    <input type="number" min="0" className="w-12 border rounded p-1 text-center" 
                           defaultValue={scores[m.id]?.away || 0}
                           onChange={e => setScores(prev => ({...prev, [m.id]: {...prev[m.id], away: parseInt(e.target.value) || 0}}))} />
                    <button onClick={() => handleSave(m.id)} className="bg-green-600 text-white px-3 py-1 rounded hover:bg-green-700 ml-2">Guardar</button>
                  </div>
                ) : (
                  <div className="flex items-center gap-4">
                    {savedScores[m.id] ? (
                      <span className="font-bold text-tm-blue bg-tm-light-blue px-3 py-1 rounded">
                        {savedScores[m.id].home} - {savedScores[m.id].away}
                      </span>
                    ) : null}
                    {m.is_locked ? (
                      <span className="text-red-500 font-bold">Bloqueado</span>
                    ) : (
                      <button 
                        onClick={() => setEditingId(m.id)}
                        className="text-tm-blue hover:text-blue-900 font-semibold underline text-sm"
                      >
                        {savedScores[m.id] ? 'Editar' : 'Pronosticar'}
                      </button>
                    )}
                  </div>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
