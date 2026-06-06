import { useState, useEffect, useMemo } from 'react';
import { useParams, Link } from 'react-router-dom';
import { apiClient } from '../shared/api/apiClient';

interface PredictionWithMatch {
  group_name: string;
  home_team: string;
  home_flag: string;
  away_team: string;
  away_flag: string;
  match_date: string;
  status: string;
  home_score_real: number;
  away_score_real: number;
  points: number;
  has_prediction: boolean;
  home_score_pred: number;
  away_score_pred: number;
}

interface UserPredictionsResponse {
  user: { id: number; username: string };
  predictions: PredictionWithMatch[];
}

const POINTS_BADGE: { [key: number]: string } = {
  3: 'bg-green-100 text-green-800',
  2: 'bg-blue-100 text-blue-800',
  1: 'bg-yellow-100 text-yellow-800',
  0: 'bg-red-100 text-red-800',
};

export function UserPredictionsPage() {
  const { id } = useParams<{ id: string }>();
  const [data, setData] = useState<UserPredictionsResponse | null>(null);
  const [activeGroup, setActiveGroup] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    apiClient
      .get(`/users/${id}/predictions`)
      .then((d: UserPredictionsResponse) => {
        setData(d);
        const first = d.predictions.find(p => p.group_name);
        if (first) setActiveGroup(first.group_name);
      })
      .catch((e: any) => {
        if (e?.message?.includes('not found') || e?.message?.includes('404')) {
          setError('Usuario no encontrado');
        } else {
          setError(e?.message || 'Error al cargar las predicciones');
        }
      });
  }, [id]);

  const availableGroups = useMemo(() => {
    if (!data) return [];
    return Array.from(new Set(data.predictions.map(p => p.group_name))).sort();
  }, [data]);

  const filtered = useMemo(() => {
    if (!data || !activeGroup) return [];
    return data.predictions.filter(p => p.group_name === activeGroup);
  }, [data, activeGroup]);

  if (error) {
    return (
      <div className="bg-white p-6 rounded-lg shadow-sm">
        <p className="text-red-600 mb-4">{error}</p>
        <Link to="/ranking" className="text-tm-blue hover:underline">← Volver al ranking</Link>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="bg-white p-6 rounded-lg shadow-sm">
        <p className="text-gray-600">Cargando predicciones...</p>
      </div>
    );
  }

  return (
    <div className="bg-white p-6 rounded-lg shadow-sm">
      <div className="flex justify-between items-center mb-4 flex-wrap gap-2">
        <h1 className="text-2xl font-bold border-b-2 border-tm-blue pb-2">
          Pronósticos de <span className="text-tm-blue">{data.user.username}</span>
        </h1>
        <Link to="/ranking" className="text-tm-blue hover:underline font-semibold">
          ← Volver al ranking
        </Link>
      </div>

      {availableGroups.length === 0 ? (
        <p className="text-gray-600 italic">Este usuario aún no hizo pronósticos.</p>
      ) : (
        <>
          <div className="flex flex-wrap gap-2 mb-4 border-b pb-3">
            {availableGroups.map(g => (
              <button
                key={g}
                onClick={() => setActiveGroup(g)}
                className={`px-3 py-1 rounded font-bold text-sm ${
                  activeGroup === g
                    ? 'bg-tm-blue text-white'
                    : 'bg-gray-100 hover:bg-gray-200 text-tm-blue'
                }`}
              >
                Grupo {g}
              </button>
            ))}
          </div>

          <div className="overflow-x-auto -mx-6 px-6">
            <table className="w-full text-left min-w-[480px]">
              <thead>
                <tr className="bg-gray-100 text-tm-blue uppercase text-sm">
                  <th className="p-2 md:p-3 text-right">Local</th>
                  <th className="p-2 md:p-3 text-center">Resultado</th>
                  <th className="p-2 md:p-3">Visitante</th>
                  <th className="p-2 md:p-3 text-center">Pronóstico</th>
                  <th className="p-2 md:p-3 text-center">Puntos</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((m, i) => {
                  const isFinished = m.status === 'finished';
                  const hasPred = m.has_prediction;
                  const resultLabel = isFinished
                    ? `${m.home_score_real} - ${m.away_score_real}`
                    : '—';
                  const predLabel = hasPred ? `${m.home_score_pred} - ${m.away_score_pred}` : '—';
                  return (
                    <tr key={`${m.group_name}-${i}`} className="border-b">
                      <td className="p-2 md:p-3 font-medium text-right whitespace-nowrap">
                        {m.home_team} <span className="text-xl ml-2">{m.home_flag}</span>
                      </td>
                      <td className="p-2 md:p-3 text-center font-bold text-tm-blue whitespace-nowrap">
                        {resultLabel}
                      </td>
                      <td className="p-2 md:p-3 font-medium whitespace-nowrap">
                        <span className="text-xl mr-2">{m.away_flag}</span> {m.away_team}
                      </td>
                      <td className="p-2 md:p-3 text-center font-bold whitespace-nowrap">
                        {predLabel}
                      </td>
                      <td className="p-2 md:p-3 text-center whitespace-nowrap">
                        {hasPred && isFinished ? (
                          <span className={`inline-block px-2 py-1 rounded font-bold text-sm ${POINTS_BADGE[m.points] || 'bg-gray-100'}`}>
                            {m.points}
                          </span>
                        ) : (
                          <span className="text-gray-400">—</span>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  );
}
