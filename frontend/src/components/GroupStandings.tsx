interface TeamStanding {
  position: number;
  team: string;
  flag: string;
  played: number;
  won: number;
  drawn: number;
  lost: number;
  goals_for: number;
  goals_against: number;
  goal_difference: number;
  points: number;
}

interface GroupStanding {
  group: string;
  teams: TeamStanding[];
}

interface GroupStandingsProps {
  groupName: string;
  standings: GroupStanding[];
}

export function GroupStandings({ groupName, standings }: GroupStandingsProps) {
  const group = standings.find((s) => s.group === groupName);

  if (!group) {
    return (
      <div className="mt-6 p-4 bg-gray-50 rounded border border-gray-200 text-sm text-gray-500 text-center">
        Cargando tabla de posiciones del grupo {groupName}...
      </div>
    );
  }

  return (
    <div className="mt-6 bg-white rounded-lg border border-gray-200 shadow-sm">
      <h2 className="text-lg font-bold text-tm-blue px-4 py-3 border-b border-gray-200">
        Tabla de posiciones - Grupo {groupName}
      </h2>
      <div className="overflow-x-auto">
        <table className="w-full text-sm text-left min-w-[600px]">
          <thead>
            <tr className="bg-gray-100 text-tm-blue uppercase text-xs tracking-wider">
              <th className="p-2 text-center">#</th>
              <th className="p-2">Equipo</th>
              <th className="p-2 text-center">PJ</th>
              <th className="p-2 text-center">G</th>
              <th className="p-2 text-center">E</th>
              <th className="p-2 text-center">P</th>
              <th className="p-2 text-center">GF</th>
              <th className="p-2 text-center">GC</th>
              <th className="p-2 text-center">Dif</th>
              <th className="p-2 text-center font-bold">Pts</th>
            </tr>
          </thead>
          <tbody>
            {group.teams.map((t) => (
              <tr key={t.team} className="border-b hover:bg-tm-light-blue transition-colors">
                <td className="p-2 text-center font-medium">{t.position}</td>
                <td className="p-2 font-medium whitespace-nowrap">
                  <span className="mr-2">{t.flag}</span>
                  {t.team}
                </td>
                <td className="p-2 text-center">{t.played}</td>
                <td className="p-2 text-center">{t.won}</td>
                <td className="p-2 text-center">{t.drawn}</td>
                <td className="p-2 text-center">{t.lost}</td>
                <td className="p-2 text-center">{t.goals_for}</td>
                <td className="p-2 text-center">{t.goals_against}</td>
                <td className="p-2 text-center">
                  {t.goal_difference > 0 ? `+${t.goal_difference}` : t.goal_difference}
                </td>
                <td className="p-2 text-center font-bold text-tm-blue">{t.points}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
