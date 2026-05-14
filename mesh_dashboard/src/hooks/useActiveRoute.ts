import { useEffect, useState } from 'react';
import axios from 'axios';

interface Route {
  source: number;
  target: number;
  path: number[];
  latency_ms: number;
  hop_count: number;
  reroute_count: number;
}

const apiUrl = process.env.REACT_APP_API_URL || 'http://localhost:8000/api/traceroute';

export const useActiveRoute = (source: number, target: number) => {
  const [route, setRoute] = useState<Route | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchRoute = async () => {
      setLoading(true);
      try {
        const response = await axios.get<Route>(`${apiUrl}/path/${source}/${target}`);
        setRoute(response.data);
        setError(null);
      } catch (err) {
        setError(String(err));
        setRoute(null);
      } finally {
        setLoading(false);
      }
    };

    const interval = setInterval(fetchRoute, 5000); // Update every 5 seconds
    fetchRoute(); // Initial fetch

    return () => clearInterval(interval);
  }, [source, target]);

  const triggerReroute = async () => {
    try {
      const response = await axios.post(`${apiUrl}/reroute/${source}/${target}`);
      if (response.data.route) {
        setRoute(response.data.route);
      }
    } catch (err) {
      setError(String(err));
    }
  };

  return { route, loading, error, triggerReroute };
};
