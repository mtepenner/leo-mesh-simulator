import React from 'react';

interface LaserLinksProps {
  connections: Array<[number, number]>;
  positions: Map<number, [number, number, number]>;
}

/**
 * LaserLinks renders the inter-satellite links as glowing lines
 * Uses WebGL for efficient rendering of thousands of connections
 */
const LaserLinks: React.FC<LaserLinksProps> = ({ connections, positions }) => {
  // Note: In a full implementation, this would use THREE.Line2 or custom shaders
  // for better performance and visual appeal

  return (
    <>
      {/* In the Three.js canvas, we render lines from the Constellation component */}
      {/* This component serves as a logical unit for managing connection rendering */}
      <group>
        {connections.slice(0, 1000).map(([from, to], i) => {
          const posFrom = positions.get(from);
          const posTo = positions.get(to);

          if (!posFrom || !posTo) return null;

          const scale = 1e-5;
          return (
            <line key={`link-${i}`}>
              <bufferGeometry>
                <bufferAttribute
                  attach="attributes-position"
                  count={2}
                  array={new Float32Array([
                    posFrom[0] * scale,
                    posFrom[1] * scale,
                    posFrom[2] * scale,
                    posTo[0] * scale,
                    posTo[1] * scale,
                    posTo[2] * scale,
                  ])}
                  itemSize={3}
                />
              </bufferGeometry>
              <lineBasicMaterial color="#00ccff" linewidth={0.5} transparent opacity={0.3} />
            </line>
          );
        })}
      </group>
    </>
  );
};

export default LaserLinks;
