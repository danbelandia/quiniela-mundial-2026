import { useState, useEffect } from 'react';
import { apiClient } from '../shared/api/apiClient';

/*
  EDUCATIONAL NOTE:
  - 'useEffect' with an empty dependency array '[]' is functionally equivalent to 
    'ngOnInit' in Angular, as it runs once after the component mounts.
  - State management here uses 'useState', which triggers re-renders. 
    In Angular, you might use 'BehaviorSubject' to stream state changes, where the 
    template subscribes using the 'async' pipe, keeping state logic outside the component.
*/

export function useMatches() {
  const [matches, setMatches] = useState([]);
  const [loading, setLoading] = useState(true);

  const fetchMatches = () => {
    apiClient.get('/matches').then(data => {
      setMatches(data);
      setLoading(false);
    });
  };

  useEffect(() => {
    fetchMatches();
  }, []);

  return { matches, loading, refresh: fetchMatches };
}

export function useSubmitPrediction() {
  const submit = async (prediction: any) => {
    return await apiClient.post('/predictions', prediction);
  };
  return { submit };
}

export function useRanking() {
  const [ranking, setRanking] = useState([]);

  useEffect(() => {
    apiClient.get('/ranking').then(data => {
      setRanking(data);
    });
  }, []);

  return { ranking };
}

export function useStandings() {
  const [standings, setStandings] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const lastFetchedRef = { current: 0 };

  const fetchStandings = () => {
    lastFetchedRef.current = Date.now();
    apiClient.get('/standings').then(data => {
      setStandings(data);
      setLoading(false);
    });
  };

  useEffect(() => {
    fetchStandings();

    const onFocus = () => {
      if (Date.now() - lastFetchedRef.current > 10000) {
        fetchStandings();
      }
    };
    window.addEventListener('focus', onFocus);
    return () => window.removeEventListener('focus', onFocus);
  }, []);

  return { standings, loading, refresh: fetchStandings };
}

export function useQualifierPrediction(groupName: string, userId: number) {
  const [prediction, setPrediction] = useState<{ predicted_first: string; predicted_second: string }>({
    predicted_first: '',
    predicted_second: '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [savedAt, setSavedAt] = useState<number>(0);
  const debounceRef = { current: 0 as number | undefined };
  const abortRef = { current: null as AbortController | null };

  useEffect(() => {
    if (!groupName || !userId) return;
    const ctrl = new AbortController();
    abortRef.current?.abort();
    abortRef.current = ctrl;

    setLoading(true);
    setError(null);
    fetch(`${(apiClient as any).BASE_URL}/groups/${encodeURIComponent(groupName)}/qualifier-predictions/me?user_id=${userId}`, {
      signal: ctrl.signal,
    })
      .then(r => r.ok ? r.json() : null)
      .then(data => {
        if (data) {
          setPrediction({
            predicted_first: data.predicted_first || '',
            predicted_second: data.predicted_second || '',
          });
        }
      })
      .catch(err => {
        if (err.name !== 'AbortError') setError(err.message);
      })
      .finally(() => setLoading(false));

    return () => ctrl.abort();
  }, [groupName, userId]);

  const save = (first: string, second: string) => {
    if (!groupName || !userId) return;
    if (debounceRef.current) window.clearTimeout(debounceRef.current);
    debounceRef.current = window.setTimeout(async () => {
      setError(null);
      try {
        const r = await apiClient.rawFetch(
          `/groups/${encodeURIComponent(groupName)}/qualifier-predictions`,
          {
            method: 'PUT',
            body: JSON.stringify({
              user_id: userId,
              predicted_first: first,
              predicted_second: second,
            }),
          }
        );
        if (!r.ok) {
          setError(await r.text());
          return;
        }
        setSavedAt(Date.now());
      } catch (e: any) {
        setError(e.message);
      }
    }, 800);
  };

  return { prediction, setPrediction, save, loading, error, savedAt };
}

export function useQualifierPredictionsByUser(userId: number) {
  const [views, setViews] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!userId) return;
    const ctrl = new AbortController();
    setLoading(true);
    setError(null);
    fetch(`${(apiClient as any).BASE_URL}/users/${userId}/qualifier-predictions`, {
      signal: ctrl.signal,
    })
      .then(r => {
        if (r.status === 404) throw new Error('Usuario no encontrado');
        if (!r.ok) throw new Error(`Error ${r.status}`);
        return r.json();
      })
      .then(data => setViews(Array.isArray(data) ? data : []))
      .catch(err => {
        if (err.name !== 'AbortError') setError(err.message);
      })
      .finally(() => setLoading(false));
    return () => ctrl.abort();
  }, [userId]);

  return { views, loading, error };
}

export interface AppConfig {
  qualifier_lock_at: string;
  match_lock_window_hours: number;
  top_scorer_candidates: TopScorerCandidate[];
}

export interface TopScorerCandidate {
  name: string;
  team: string;
  flag: string;
}

export interface TopScorerPrediction {
  id: number;
  user_id: number;
  predicted_player: string;
  created_at: string;
  updated_at: string;
}

export function useConfig() {
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const ctrl = new AbortController();
    fetch(`${(apiClient as any).BASE_URL}/config`, { signal: ctrl.signal })
      .then(r => r.ok ? r.json() : null)
      .then(data => {
        if (data) setConfig(data);
      })
      .catch(err => {
        if (err.name !== 'AbortError') console.error('useConfig error:', err);
      })
      .finally(() => setLoading(false));
    return () => ctrl.abort();
  }, []);

  return { config, loading };
}

export function useMyTopScorerPrediction(userId: number) {
  const [prediction, setPrediction] = useState<TopScorerPrediction | null>(null);
  const [loading, setLoading] = useState(false);

  const load = () => {
    if (!userId) return;
    setLoading(true);
    fetch(`${(apiClient as any).BASE_URL}/top-scorer-prediction/me?user_id=${userId}`)
      .then(r => r.ok ? r.json() : null)
      .then(data => setPrediction(data || null))
      .catch(() => setPrediction(null))
      .finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, [userId]);

  return { prediction, loading, refresh: load };
}

export function useUpsertTopScorerPrediction() {
  const upsert = async (userId: number, predictedPlayer: string) => {
    const r = await apiClient.rawFetch('/top-scorer-prediction/me', {
      method: 'PUT',
      body: JSON.stringify({ user_id: userId, predicted_player: predictedPlayer }),
    });
    if (!r.ok) {
      const text = await r.text();
      throw new Error(text || `Error ${r.status}`);
    }
    return r.json();
  };
  return { upsert };
}

export function useUserTopScorerPrediction(userId: number) {
  const [prediction, setPrediction] = useState<TopScorerPrediction | null>(null);

  useEffect(() => {
    if (!userId) return;
    const ctrl = new AbortController();
    fetch(`${(apiClient as any).BASE_URL}/users/${userId}/top-scorer-prediction`, { signal: ctrl.signal })
      .then(r => r.ok ? r.json() : null)
      .then(data => setPrediction(data || null))
      .catch(err => { if (err.name !== 'AbortError') setPrediction(null); });
    return () => ctrl.abort();
  }, [userId]);

  return { prediction };
}