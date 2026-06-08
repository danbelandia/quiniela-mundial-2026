import { useState, useMemo } from 'react';
import { useConfig, useMyTopScorerPrediction, useUpsertTopScorerPrediction } from '../features/hooks';
import { useAuth } from '../shared/AuthContext';

export function GoleadorPage() {
  const { config, loading: configLoading } = useConfig();
  const currentUserId = parseInt(localStorage.getItem('userId') || '0');
  const { prediction, refresh } = useMyTopScorerPrediction(currentUserId);
  const { upsert } = useUpsertTopScorerPrediction();
  const [selected, setSelected] = useState<string>(prediction?.predicted_player || '');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [savedAt, setSavedAt] = useState<number>(0);

  const isLocked = useMemo(() => {
    if (!config?.qualifier_lock_at) return false;
    return Date.now() >= new Date(config.qualifier_lock_at).getTime();
  }, [config]);

  if (configLoading) return <div className="text-center mt-10">Cargando...</div>;

  const candidates = config?.top_scorer_candidates || [];
  const candidatesByTeam = candidates.reduce((acc, c) => {
    if (!acc[c.team]) acc[c.team] = [];
    acc[c.team].push(c);
    return acc;
  }, {} as Record<string, typeof candidates>);

  const handleSave = async () => {
    if (!selected) return;
    setError(null);
    setSaving(true);
    try {
      await upsert(currentUserId, selected);
      await refresh();
      setSavedAt(Date.now());
    } catch (e: any) {
      setError(e.message);
    } finally {
      setSaving(false);
    }
  };

  const currentCandidate = candidates.find(c => c.name === prediction?.predicted_player);

  return (
    <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-100">
      <h1 className="text-2xl font-bold text-tm-blue mb-2">Goleador de la fase de grupos</h1>
      <p className="text-sm text-gray-600 mb-4">
        Pronosticá quién será el máximo anotador de la fase de grupos. Si acertás, sumás <strong>6 puntos</strong>.
      </p>

      {isLocked && (
        <div className="mb-4 p-3 bg-yellow-50 border border-yellow-300 rounded text-sm text-yellow-800">
          El plazo para pronosticar al goleador cerró el <strong>11/06 a las 14:00 CLT</strong>. No se pueden hacer más cambios.
        </div>
      )}

      {currentCandidate && (
        <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded">
          <div className="text-xs text-gray-600 mb-1">Tu pick actual:</div>
          <div className="flex items-center gap-2">
            <span className="text-2xl">{currentCandidate.flag}</span>
            <div>
              <div className="font-bold">{currentCandidate.name}</div>
              <div className="text-xs text-gray-500">{currentCandidate.team}</div>
            </div>
          </div>
        </div>
      )}

      {!isLocked && (
        <>
          <label className="block text-sm font-medium text-gray-700 mb-2">Elegí un jugador:</label>
          <select
            value={selected}
            onChange={e => setSelected(e.target.value)}
            className="w-full border border-gray-300 rounded p-2 mb-3"
            disabled={saving}
          >
            <option value="">— Seleccionar —</option>
            {Object.entries(candidatesByTeam).map(([team, players]) => (
              <optgroup key={team} label={team}>
                {players.map(p => (
                  <option key={p.name} value={p.name}>
                    {p.flag} {p.name}
                  </option>
                ))}
              </optgroup>
            ))}
          </select>

          <button
            onClick={handleSave}
            disabled={!selected || saving}
            className="bg-tm-blue text-white px-4 py-2 rounded font-bold hover:bg-blue-900 disabled:opacity-50"
          >
            {saving ? 'Guardando...' : (currentCandidate ? 'Cambiar pick' : 'Guardar pick')}
          </button>

          {error && (
            <p className="mt-3 text-sm text-red-600">{error}</p>
          )}
          {savedAt > 0 && !error && (
            <p className="mt-3 text-sm text-green-600">✓ Guardado</p>
          )}
        </>
      )}
    </div>
  );
}
