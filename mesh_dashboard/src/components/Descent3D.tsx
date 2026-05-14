import React, { useEffect, useRef } from 'react';
import { Canvas, useThree, useFrame } from '@react-three/fiber';
import { PerspectiveCamera, OrbitControls, Stars } from '@react-three/drei';
import * as THREE from 'three';

interface Satellite {
  id: number;
  position: [number, number, number];
  velocity: [number, number, number];
}

interface Constellation3DProps {
  satellites: Satellite[];
  connections: Array<[number, number]>;
}

const Constellation3D: React.FC<Constellation3DProps> = ({ satellites, connections }) => {
  const earthMeshRef = useRef<THREE.Mesh>(null);
  const satelliteMeshRef = useRef<THREE.InstancedMesh>(null);
  const lineRef = useRef<THREE.LineSegments>(null);

  // Normalize position for visualization (scale down to prevent huge coordinates)
  const normalizePosition = (pos: [number, number, number]): [number, number, number] => {
    const scale = 1e-5; // Divide by 100,000 for visualization
    return [pos[0] * scale, pos[1] * scale, pos[2] * scale];
  };

  // Draw Earth
  useFrame(() => {
    if (earthMeshRef.current) {
      earthMeshRef.current.rotation.y += 0.0001;
    }
  });

  // Update satellite positions (instanced rendering)
  useFrame(() => {
    if (satelliteMeshRef.current && satellites.length > 0) {
      const dummy = new THREE.Object3D();
      satellites.forEach((sat, i) => {
        if (i < satelliteMeshRef.current!.count) {
          const [x, y, z] = normalizePosition(sat.position);
          dummy.position.set(x, y, z);
          dummy.scale.set(0.1, 0.1, 0.1);
          dummy.updateMatrix();
          satelliteMeshRef.current!.setMatrixAt(i, dummy.matrix);
        }
      });
      satelliteMeshRef.current.instanceMatrix.needsUpdate = true;
    }
  });

  // Build connection geometry
  useEffect(() => {
    if (lineRef.current && connections.length > 0) {
      const positions: number[] = [];

      connections.forEach(([from, to]) => {
        const satFrom = satellites.find(s => s.id === from);
        const satTo = satellites.find(s => s.id === to);

        if (satFrom && satTo) {
          const [x1, y1, z1] = normalizePosition(satFrom.position);
          const [x2, y2, z2] = normalizePosition(satTo.position);

          positions.push(x1, y1, z1, x2, y2, z2);
        }
      });

      if (positions.length > 0) {
        const geometry = new THREE.BufferGeometry();
        geometry.setAttribute('position', new THREE.BufferAttribute(new Float32Array(positions), 3));
        lineRef.current.geometry = geometry;
      }
    }
  }, [connections, satellites]);

  return (
    <>
      <PerspectiveCamera makeDefault position={[0, 0, 100]} />
      <OrbitControls autoRotate autoRotateSpeed={0.5} />
      <Stars radius={1000} depth={50} count={5000} factor={4} />

      {/* Earth */}
      <mesh ref={earthMeshRef}>
        <sphereGeometry args={[6.371, 64, 64]} />
        <meshPhongMaterial color="#1a5f7a" emissive="#0d2a3a" />
      </mesh>

      {/* Satellites as instanced mesh */}
      <instancedMesh
        ref={satelliteMeshRef}
        args={[
          new THREE.SphereGeometry(1, 8, 8),
          new THREE.MeshBasicMaterial({ color: '#00ff00' }),
          Math.max(500, satellites.length),
        ]}
      />

      {/* Connection lines */}
      <lineSegments ref={lineRef}>
        <bufferGeometry />
        <lineBasicMaterial color="#00ccff" linewidth={1} transparent opacity={0.3} />
      </lineSegments>

      {/* Lighting */}
      <ambientLight intensity={0.4} />
      <directionalLight position={[10, 10, 5]} intensity={0.8} />
    </>
  );
};

export default Constellation3D;
