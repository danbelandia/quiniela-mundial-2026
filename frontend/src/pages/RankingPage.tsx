import { useState, useEffect } from 'react';

const USERS_PER_PAGE = 15;

const MEDALS: { [key: string]: string } = {
  '1': '🥇',
  '2': '🥈',
  '3': '🥉',
};

const PODIUM_BG: { [key: string]: string } = {
  '1': 'bg-yellow-100',
  '2': 'bg-gray-200',
  '3': 'bg-amber-200',
};

export function RankingPage() {
  const [users, setUsers] = useState<any[]>([]);
  const [currentPage, setCurrentPage] = useState(1);

  useEffect(() => {
    fetch('http://localhost:8080/ranking')
      .then(res => res.json())
      .then(data => {
        const sorted = data.sort((a: any, b: any) => b.score - a.score);
        setUsers(sorted);
      });
  }, []);

  const totalPages = Math.max(1, Math.ceil(users.length / USERS_PER_PAGE));
  const paginatedUsers = users.slice(
    (currentPage - 1) * USERS_PER_PAGE,
    currentPage * USERS_PER_PAGE
  );

  useEffect(() => {
    if (currentPage > totalPages) setCurrentPage(1);
  }, [totalPages, currentPage]);

  return (
    <div className="bg-white p-6 rounded-lg shadow-sm">
      <h1 className="text-2xl font-bold border-b-2 border-tm-blue pb-2 mb-6">Ranking de Usuarios</h1>
      <table className="w-full text-left">
        <thead>
          <tr className="bg-gray-100 text-tm-blue uppercase text-sm">
            <th className="p-3">Pos.</th>
            <th className="p-3">Usuario</th>
            <th className="p-3">Email</th>
            <th className="p-3">Puntos</th>
          </tr>
        </thead>
        <tbody>
          {paginatedUsers.map((u, i) => {
            const pos = (currentPage - 1) * USERS_PER_PAGE + i + 1;
            const medal = MEDALS[String(pos)];
            const podiumBg = PODIUM_BG[String(pos)] ?? '';
            return (
              <tr key={u.id} className={`border-b ${podiumBg}`}>
                <td className="p-3 text-2xl">{medal ? medal : pos}</td>
                <td className="p-3 font-bold text-tm-blue">{u.username}</td>
                <td className="p-3">{u.email}</td>
                <td className="p-3 font-bold text-tm-blue">{u.score}</td>
              </tr>
            );
          })}
        </tbody>
      </table>

      <div className="flex justify-between items-center mt-4">
        <span className="text-sm text-gray-600">
          Página {currentPage} de {totalPages} ({users.length} usuarios)
        </span>
        <div className="flex gap-2">
          <button
            onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
            disabled={currentPage === 1}
            className="px-3 py-1 rounded border bg-gray-100 hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            ← Anterior
          </button>
          {Array.from({ length: totalPages }, (_, i) => i + 1).map(page => (
            <button
              key={page}
              onClick={() => setCurrentPage(page)}
              className={`px-3 py-1 rounded border ${currentPage === page ? 'bg-tm-blue text-white' : 'bg-gray-100 hover:bg-gray-200'}`}
            >
              {page}
            </button>
          ))}
          <button
            onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
            disabled={currentPage === totalPages}
            className="px-3 py-1 rounded border bg-gray-100 hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Siguiente →
          </button>
        </div>
      </div>

      <div className="mt-8 border-t-2 border-tm-blue pt-6">
        <h2 className="text-xl font-bold text-tm-blue mb-4">¿Cómo se calculan los puntos?</h2>
        <table className="w-full text-left">
          <thead>
            <tr className="bg-gray-100 text-tm-blue uppercase text-sm">
              <th className="p-3">Condición</th>
              <th className="p-3 text-center">Puntos</th>
            </tr>
          </thead>
          <tbody>
            <tr className="border-b">
              <td className="p-3">Marcador exacto (incluye empates exactos)</td>
              <td className="p-3 text-center font-bold text-green-600">3</td>
            </tr>
            <tr className="border-b">
              <td className="p-3">Ganador correcto, marcador incorrecto</td>
              <td className="p-3 text-center font-bold text-tm-blue">2</td>
            </tr>
            <tr className="border-b">
              <td className="p-3">Empate pronosticado, resultado fue empate, marcador incorrecto</td>
              <td className="p-3 text-center font-bold text-yellow-600">1</td>
            </tr>
            <tr className="border-b">
              <td className="p-3">Cualquier otro caso</td>
              <td className="p-3 text-center font-bold text-red-600">0</td>
            </tr>
          </tbody>
        </table>
        <p className="text-xs text-gray-500 mt-3 italic">
          "Ganador correcto" se determina comparando el signo de (goles local - goles visitante) del pronóstico con el del resultado real. El empate exacto ya queda cubierto por la regla de 3 puntos.
        </p>
      </div>
    </div>
  );
}