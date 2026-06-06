import { useEffect, useMemo } from 'react';
import { useQualifierPrediction } from '../features/hooks';

interface Props {
  groupName: string;
  userId: number;
  isLocked: boolean;
  teamsInGroup: { name: string; flag: string }[];
}

export function QualifierPredictionSection({ groupName, userId, isLocked, teamsInGroup }: Props) {
  const { prediction, setPrediction, save, error, savedAt } = useQualifierPrediction(groupName, userId);

  const sortedTeams = useMemo(
    () => [...teamsInGroup].sort((a, b) => a.name.localeCompare(b.name)),
    [teamsInGroup]
  );

  useEffect(() => {
    if (prediction.predicted_first && prediction.predicted_second) {
      save(prediction.predicted_first, prediction.predicted_second);
    }
  }, [prediction.predicted_first, prediction.predicted_second]);

  const firstOptions = sortedTeams;
  const secondOptions = sortedTeams.filter(t => t.name !== prediction.predicted_first);

  return (
    <div className="mt-6 p-4 bg-tm-light-blue/30 rounded-lg border border-tm-blue/20">
      <h2 className="text-lg font-bold text-tm-blue mb-2">
        Predicción 1° y 2° del grupo
      </h2>
      <p className="text-sm text-gray-600 mb-3">
        Sumás 3 puntos por cada acierto. Se otorgan solo cuando los 6 partidos del grupo finalizan.
      </p>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-semibold text-gray-700 mb-1">1°</label>
          <select
            disabled={isLocked}
            value={prediction.predicted_first}
            onChange={e => setPrediction(prev => ({
              ...prev,
              predicted_first: e.target.value,
              predicted_second: prev.predicted_second === e.target.value ? '' : prev.predicted_second,
            }))}
            className="w-full border border-gray-300 rounded p-2 bg-white disabled:bg-gray-100 disabled:cursor-not-allowed"
          >
            <option value="">— Seleccioná —</option>
            {firstOptions.map(t => (
              <option key={t.name} value={t.name}>{t.flag} {t.name}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-sm font-semibold text-gray-700 mb-1">2°</label>
          <select
            disabled={isLocked || !prediction.predicted_first}
            value={prediction.predicted_second}
            onChange={e => setPrediction(prev => ({ ...prev, predicted_second: e.target.value }))}
            className="w-full border border-gray-300 rounded p-2 bg-white disabled:bg-gray-100 disabled:cursor-not-allowed"
          >
            <option value="">— Seleccioná —</option>
            {secondOptions.map(t => (
              <option key={t.name} value={t.name}>{t.flag} {t.name}</option>
            ))}
          </select>
        </div>
      </div>

      <div className="mt-3 min-h-[1.5rem] text-sm">
        {isLocked && (
          <span className="text-red-600 font-semibold">
            🔒 Los pronósticos de este grupo están bloqueados
          </span>
        )}
        {!isLocked && error && (
          <span className="text-red-600">{error}</span>
        )}
        {!isLocked && !error && savedAt > 0 && (
          <span className="text-green-600">✓ Guardado</span>
        )}
      </div>
    </div>
  );
}
