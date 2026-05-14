import React, { useState, useEffect } from 'react';
import { Canvas } from '@react-three/fiber';
import Constellation from './components/Constellation';
import RoutingHUD from './components/RoutingHUD';
import { useNetworkState } from './hooks/useNetworkState';
import { useActiveRoute } from './hooks/useActiveRoute';
import './App.css';

interface Satellite {
  id: number;
  position: [number, number, number];
  velocity: [number, number, number];
}

const App: React.FC = () => {
  const [satellites, setSatellites] = useState<Satellite[]>([]);
  const [sourceId, setSourceId] = useState(10);
  const [targetId, setTargetId] = useState(100);

  const networkState = useNetworkState();
  const route = useActiveRoute(sourceId, targetId);

  // Simulate satellite positions for demo (in real app, would fetch from API)
  useEffect(() => {
    const generateSatellites = () => {
      const sats: Satellite[] = [];
      const earthRadius = 6.371e6;
      const leoAltitude = 550e3;
      const radius = earthRadius + leoAltitude;

      for (let i = 0; i < 500; i++) {
        const theta = Math.random() * Math.PI * 2;
        const phi = Math.random() * Math.PI;

        const x = radius * Math.sin(phi) * Math.cos(theta);
        const y = radius * Math.sin(phi) * Math.sin(theta);
        const z = radius * Math.cos(phi);

        sats.push({
          id: i,
          position: [x, y, z],
          velocity: [0, 0, 0],
        });
      }

      setSatellites(sats);
    };

    generateSatellites();
  }, []);

  return (
    <div className="app">
      <div className="header">
        <h1>LEO Mesh Network Operations Center</h1>
        <div className="network-stats">
          {networkState.state && (
            <>
              <span>{networkState.state.nodes} Satellites</span>
              <span>{networkState.state.edges} Links</span>
              <span>Status: {networkState.state.is_connected ? '🟢 Connected' : '🔴 Offline'}</span>
            </>
          )}
        </div>
      </div>

      <div className="main-panel">
        <div className="canvas-container">
          <Canvas>
            <Constellation satellites={satellites} connectedPairs={networkState.state?.edges || 0} />
          </Canvas>
        </div>

        <div className="control-panel">
          <div className="control-section">
            <label>Source Satellite ID:</label>
            <input
              type="number"
              value={sourceId}
              onChange={e => setSourceId(parseInt(e.target.value))}
              min="0"
              max="499"
            />
          </div>

          <div className="control-section">
            <label>Target Satellite ID:</label>
            <input
              type="number"
              value={targetId}
              onChange={e => setTargetId(parseInt(e.target.value))}
              min="0"
              max="499"
            />
          </div>

          {route.route && (
            <div className="route-info">
              <h3>Active Route</h3>
              <p>
                <strong>Path:</strong> {route.route.path.join(' → ')}
              </p>
              <p>
                <strong>Latency:</strong> {route.route.latency_ms.toFixed(2)} ms
              </p>
              <p>
                <strong>Hops:</strong> {route.route.hop_count}
              </p>
              <button onClick={route.triggerReroute} disabled={route.loading}>
                {route.loading ? 'Checking...' : 'Force Reroute'}
              </button>
            </div>
          )}

          {route.error && <div className="error-message">{route.error}</div>}
        </div>
      </div>

      {route.route && (
        <RoutingHUD
          source={route.route.source}
          target={route.route.target}
          path={route.route.path}
          latencyMs={route.route.latency_ms}
          hopCount={route.route.hop_count}
          rerouteCount={route.route.reroute_count}
        />
      )}
    </div>
  );
};

export default App;
