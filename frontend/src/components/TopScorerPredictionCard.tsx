import type { TopScorerCandidate } from '../features/hooks';

export interface TopScorerPredictionCardProps {
  candidate: TopScorerCandidate | null;
  actualPlayer: string | null;
  points: number;
}

export function TopScorerPredictionCard({ candidate, actualPlayer, points }: TopScorerPredictionCardProps) {
  if (!candidate) {
    return (
      <div className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm">
        <h3 className="text-lg font-bold text-tm-blue mb-2">Goleador fase de grupos</h3>
        <p className="text-gray-500 italic">Aún no pronosticó</p>
      </div>
    );
  }

  const isCorrect = actualPlayer !== null && candidate.name === actualPlayer;
  const isClosed = actualPlayer !== null;

  return (
    <div className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm">
      <div className="flex justify-between items-start mb-2">
        <h3 className="text-lg font-bold text-tm-blue">Goleador fase de grupos</h3>
        {isClosed && (
          <span className={`px-2 py-1 rounded text-xs font-bold ${
            isCorrect ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
          }`}>
            {isCorrect ? `+${points} pts` : 'No acertó'}
          </span>
        )}
      </div>
      <div className="flex items-center gap-3">
        <span className="text-3xl">{candidate.flag}</span>
        <div>
          <div className="font-bold text-gray-800 text-lg">{candidate.name}</div>
          <div className="text-sm text-gray-500">{candidate.team}</div>
        </div>
      </div>
      {isClosed && !isCorrect && actualPlayer && (
        <p className="text-sm text-gray-500 mt-2">Resultado real: <strong>{actualPlayer}</strong></p>
      )}
    </div>
  );
}
