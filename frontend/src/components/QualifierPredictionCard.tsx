export interface QualifierPredictionView {
  group_name: string;
  predicted_first: string;
  predicted_first_flag: string;
  predicted_second: string;
  predicted_second_flag: string;
  actual_first: string;
  actual_first_flag: string;
  actual_second: string;
  actual_second_flag: string;
  points_earned: number;
  group_closed: boolean;
}

export function QualifierPredictionCard({ view }: { view: QualifierPredictionView }) {
  if (!view.predicted_first) {
    return (
      <div className="mb-4 p-3 rounded bg-gray-50 border border-gray-200 text-gray-500 text-sm italic">
        Sin pron&oacute;stico de clasificaci&oacute;n
      </div>
    );
  }

  const firstCorrect = view.group_closed && view.actual_first === view.predicted_first;
  const secondCorrect = view.group_closed && view.actual_second === view.predicted_second;

  return (
    <div className="mb-4 p-4 rounded border border-tm-blue/30 bg-blue-50/40">
      <div className="text-xs font-bold uppercase tracking-wide text-tm-blue mb-3">
        Pron&oacute;stico de clasificaci&oacute;n
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <PickBadge
          position="1&deg;"
          team={view.predicted_first}
          flag={view.predicted_first_flag}
          groupClosed={view.group_closed}
          correct={firstCorrect}
        />
        <PickBadge
          position="2&deg;"
          team={view.predicted_second}
          flag={view.predicted_second_flag}
          groupClosed={view.group_closed}
          correct={secondCorrect}
        />
        {view.group_closed && (
          <span className="ml-auto inline-block px-3 py-1 rounded-full font-bold text-sm bg-tm-blue text-white">
            {view.points_earned} pts
          </span>
        )}
      </div>
      {view.group_closed && (
        <div className="mt-3 pt-3 border-t border-tm-blue/20 text-sm text-gray-700">
          <span className="font-semibold text-tm-blue">Clasificaron:</span>
          <span className="ml-3">
            <span className="text-lg mr-1">{view.actual_first_flag}</span>
            {view.actual_first}
          </span>
          <span className="mx-2 text-gray-400">&middot;</span>
          <span>
            <span className="text-lg mr-1">{view.actual_second_flag}</span>
            {view.actual_second}
          </span>
        </div>
      )}
    </div>
  );
}

function PickBadge({
  position,
  team,
  flag,
  groupClosed,
  correct,
}: {
  position: string;
  team: string;
  flag: string;
  groupClosed: boolean;
  correct: boolean;
}) {
  const closedClasses = groupClosed
    ? correct
      ? 'bg-green-100 text-green-800 border border-green-300'
      : 'bg-red-50 text-red-700 border border-red-200'
    : 'bg-white border border-gray-200';

  return (
    <div className={`flex items-center gap-2 px-3 py-1.5 rounded ${closedClasses}`}>
      <span className="text-xs font-bold text-gray-500 w-4">{position}</span>
      <span className="text-xl leading-none">{flag || '🏳️'}</span>
      <span className="font-semibold text-sm">{team}</span>
      {groupClosed && (
        <span className={`text-base font-bold ${correct ? 'text-green-600' : 'text-red-500'}`}>
          {correct ? '✓' : '✗'}
        </span>
      )}
    </div>
  );
}
