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