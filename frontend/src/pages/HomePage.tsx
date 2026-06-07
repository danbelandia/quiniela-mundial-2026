import { useState, useMemo, useEffect } from 'react';
import { useMatches, useSubmitPrediction, useStandings, useConfig } from '../features/hooks';
import { useAuth } from '../shared/AuthContext';
import { apiClient } from '../shared/api/apiClient';
import { GroupStandings } from '../components/GroupStandings';
import { QualifierPredictionSection } from '../components/QualifierPredictionSection';

export function HomePage() {
  const { matches, loading, refresh } = useMatches();
  const { submit } = useSubmitPrediction();
  const { standings } = useStandings();
  const { config } = useConfig();
  const { isAdmin } = useAuth();
  const [currentGroup, setCurrentGroup] = useState('A');
  const [editingId, setEditingId] = useState<number | null>(null);
  const [scores, setScores] = useState<{ [key: number]: { home: number; away: number } }>({});
  const [savedScores, setSavedScores] = useState<{ [key: number]: { home: number; away: number } }>({});

  useEffect(() => {
    apiClient.get('/predictions').then(data => {
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

  const groupIsLocked = useMemo(() => {
    if (!config?.qualifier_lock_at) return false;
    return Date.now() >= new Date(config.qualifier_lock_at).getTime();
  }, [config]);

  const teamsInGroup = useMemo(() => {
    const map = new Map<string, string>();
    for (const m of filteredMatches) {
      if (!map.has(m.home_team)) map.set(m.home_team, m.home_flag);
      if (!map.has(m.away_team)) map.set(m.away_team, m.away_flag);
    }
    return Array.from(map.entries()).map(([name, flag]) => ({ name, flag }));
  }, [filteredMatches]);

  const currentUserId = parseInt(localStorage.getItem('userId') || '1');

  const handleSave = async (matchId: number) => {
    const s = scores[matchId] || { home: 0, away: 0 };
    const userId = localStorage.getItem('userId') || '1';

    try {
      await apiClient.post('/predictions', {
        match_id: matchId,
        home_score: s.home,
        away_score: s.away,
        user_id: parseInt(userId)
      });
      setSavedScores(prev => ({...prev, [matchId]: s}));
      setEditingId(null);
    } catch (error) {
      console.error('Fallo al guardar:', error);
      alert('Error al guardar la predicción');
    }
  };

  const handleLockAll = async (locked: boolean) => {
    try {
      await apiClient.post('/admin/matches/lock-all', { is_locked: locked });
      alert('Todos los partidos han sido ' + (locked ? 'bloqueados' : 'desbloqueados'));
      refresh();
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
      
      <div className="overflow-x-auto -mx-6 px-6">
        <table className="w-full text-left min-w-[520px]">
          <thead>
            <tr className="bg-gray-100 text-tm-blue uppercase text-xs tracking-wider">
              <th className="p-3">Fecha</th>
              <th className="p-3">Local</th>
              <th className="p-3 text-center">Resultado</th>
              <th className="p-3">Visitante</th>
              <th className="p-3 text-right">Acción</th>
            </tr>
          </thead>
          <tbody>
            {filteredMatches.map((m: any) => {
              const d = new Date(m.date);
              const datePart = d.toLocaleDateString('es-CL', { day: '2-digit', month: '2-digit', timeZone: 'America/Santiago' });
              const timePart = d.toLocaleTimeString('es-CL', { hour: 'numeric', minute: '2-digit', hour12: true, timeZone: 'America/Santiago' });
              return (
                <tr key={m.id} className="border-b hover:bg-tm-light-blue transition-colors">
                  <td className="p-3 text-sm text-gray-600 whitespace-nowrap">
                    {datePart} - {timePart}
                  </td>
                  <td className="p-3 font-medium text-gray-800 whitespace-nowrap text-right">
                    {m.home_team} <span className="text-xl ml-2">{m.home_flag}</span>
                  </td>
                  <td className="p-3 text-center">
                    {editingId === m.id ? (
                      <div className="flex gap-2 items-center justify-center">
                        <input
                          type="number"
                          min="0"
                          inputMode="numeric"
                          className="w-12 border rounded p-1 text-center"
                          value={scores[m.id]?.home ?? 0}
                          onChange={e => {
                            const raw = e.target.value;
                            const parsed = raw === '' ? 0 : parseInt(raw, 10);
                            setScores(prev => ({
                              ...prev,
                              [m.id]: { home: isNaN(parsed) ? 0 : parsed, away: prev[m.id]?.away ?? 0 }
                            }));
                          }}
                        />
                        <span className="text-gray-400 font-bold">-</span>
                        <input
                          type="number"
                          min="0"
                          inputMode="numeric"
                          className="w-12 border rounded p-1 text-center"
                          value={scores[m.id]?.away ?? 0}
                          onChange={e => {
                            const raw = e.target.value;
                            const parsed = raw === '' ? 0 : parseInt(raw, 10);
                            setScores(prev => ({
                              ...prev,
                              [m.id]: { home: prev[m.id]?.home ?? 0, away: isNaN(parsed) ? 0 : parsed }
                            }));
                          }}
                        />
                      </div>
                    ) : (
                      <span className="font-bold text-tm-blue bg-tm-light-blue px-3 py-1 rounded inline-block min-w-[60px]">
                        {savedScores[m.id] ? `${savedScores[m.id].home} - ${savedScores[m.id].away}` : '-'}
                      </span>
                    )}
                  </td>
                  <td className="p-3 font-medium text-gray-800 whitespace-nowrap">
                    <span className="text-xl mr-2">{m.away_flag}</span> {m.away_team}
                  </td>
                  <td className="p-3 text-right whitespace-nowrap">
                    {editingId === m.id ? (
                      <button onClick={() => handleSave(m.id)} className="bg-green-600 text-white px-3 py-1 rounded hover:bg-green-700">Guardar</button>
                    ) : m.is_locked ? (
                      <span className="text-red-500 font-bold">Bloqueado</span>
                    ) : (
                      <button
                        onClick={() => setEditingId(m.id)}
                        className="text-tm-blue hover:text-blue-900 font-semibold underline text-sm"
                      >
                        {savedScores[m.id] ? 'Editar' : 'Pronosticar'}
                      </button>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <GroupStandings groupName={currentGroup} standings={standings} />

      <QualifierPredictionSection
        groupName={currentGroup}
        userId={currentUserId}
        isLocked={groupIsLocked}
        teamsInGroup={teamsInGroup}
      />
    </div>
  );
}
