import { useEffect, useState } from 'react';

interface NetworkState {
  nodes: number;
  edges: number;
  is_connected: boolean;
  average_degree: number;
}

export const useNetworkState = (wsUrl: string = 'ws://localhost:8000/ws/topology') => {
  const [state, setState] = useState<NetworkState | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let ws: WebSocket | null = null;
    let reconnectTimeout: NodeJS.Timeout;

    const connect = () => {
      try {
        ws = new WebSocket(wsUrl);

        ws.onopen = () => {
          setIsConnected(true);
          setError(null);
          console.log('WebSocket connected');
        };

        ws.onmessage = event => {
          try {
            const message = JSON.parse(event.data);
            if (message.type === 'stats') {
              setState(message.data);
            }
          } catch (err) {
            console.error('Error parsing message:', err);
          }
        };

        ws.onerror = () => {
          setError('WebSocket error');
        };

        ws.onclose = () => {
          setIsConnected(false);
          // Reconnect after 3 seconds
          reconnectTimeout = setTimeout(connect, 3000);
        };
      } catch (err) {
        setError(String(err));
        reconnectTimeout = setTimeout(connect, 3000);
      }
    };

    connect();

    return () => {
      clearTimeout(reconnectTimeout);
      if (ws) {
        ws.close();
      }
    };
  }, [wsUrl]);

  return { state, isConnected, error };
};
