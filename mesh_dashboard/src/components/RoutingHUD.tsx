import React, { useState, useEffect } from 'react';
import './RoutingHUD.css';

interface RoutingHUDProps {
  source: number;
  target: number;
  path: number[];
  latencyMs: number;
  hopCount: number;
  rerouteCount: number;
}

const RoutingHUD: React.FC<RoutingHUDProps> = ({
  source,
  target,
  path,
  latencyMs,
  hopCount,
  rerouteCount,
}) => {
  const [animationFrame, setAnimationFrame] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => {
      setAnimationFrame(f => (f + 1) % 60);
    }, 16);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="routing-hud">
      <div className="hud-section metrics">
        <div className="metric">
          <label>Latency</label>
          <value className="latency">{latencyMs.toFixed(2)} ms</value>
        </div>
        <div className="metric">
          <label>Hops</label>
          <value>{hopCount}</value>
        </div>
        <div className="metric">
          <label>Reroutes</label>
          <value>{rerouteCount}</value>
        </div>
      </div>

      <div className="hud-section route">
        <div className="route-label">Active Route</div>
        <div className="route-path">
          {path.map((satId, i) => (
            <React.Fragment key={i}>
              <span className="satellite-id">{satId}</span>
              {i < path.length - 1 && <span className="arrow">→</span>}
            </React.Fragment>
          ))}
        </div>
      </div>

      <div className="hud-section pulse" style={{ opacity: 0.3 + 0.7 * Math.sin(animationFrame / 60 * Math.PI) }}>
        ○ Signal Active
      </div>
    </div>
  );
};

export default RoutingHUD;
