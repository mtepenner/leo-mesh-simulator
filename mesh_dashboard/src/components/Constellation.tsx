import React, { useEffect, useRef } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { PerspectiveCamera, OrbitControls } from '@react-three/drei';
import * as THREE from 'three';

interface ConstellationProps {
  satellites: Array<{ id: number; position: [number, number, number] }>;
  connectedPairs: number;
}

const ConstellationScene: React.FC<ConstellationProps> = ({ satellites, connectedPairs }) => {
  const instancedMeshRef = useRef<THREE.InstancedMesh>(null);

  useFrame(() => {
    if (instancedMeshRef.current) {
      const dummy = new THREE.Object3D();

      satellites.forEach((sat, i) => {
        if (i < instancedMeshRef.current!.count) {
          // Scale positions for visualization (1e-5 = shrink by 100,000x)
          const scale = 1e-5;
          dummy.position.set(sat.position[0] * scale, sat.position[1] * scale, sat.position[2] * scale);
          dummy.scale.set(0.15, 0.15, 0.15);
          dummy.updateMatrix();
          instancedMeshRef.current!.setMatrixAt(i, dummy.matrix);
        }
      });

      instancedMeshRef.current.instanceMatrix.needsUpdate = true;
    }
  });

  return (
    <>
      <PerspectiveCamera makeDefault position={[0, 0, 150]} />
      <OrbitControls autoRotate autoRotateSpeed={0.3} />

      {/* Earth */}
      <mesh>
        <sphereGeometry args={[6.371, 64, 64]} />
        <meshPhongMaterial color="#1a5f7a" emissive="#0d2a3a" />
      </mesh>

      {/* Satellites */}
      <instancedMesh
        ref={instancedMeshRef}
        args={[
          new THREE.OctahedronGeometry(1, 0),
          new THREE.MeshBasicMaterial({ color: '#00ff00', emissive: '#00aa00' }),
          Math.max(satellites.length, 500),
        ]}
      />

      {/* Lighting */}
      <ambientLight intensity={0.5} />
      <directionalLight position={[20, 20, 10]} intensity={0.8} />
      <pointLight position={[0, 0, 50]} intensity={0.3} />
    </>
  );
};

const Constellation: React.FC<ConstellationProps> = (props) => {
  return (
    <div style={{ width: '100%', height: '100%' }}>
      <Canvas shadows>
        <ConstellationScene {...props} />
      </Canvas>
    </div>
  );
};

export default Constellation;
