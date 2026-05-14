package main

import (
	"flag"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"leo-mesh-simulator/internal/kinematics"
	"leo-mesh-simulator/internal/publisher"
	"leo-mesh-simulator/internal/topology"
)

func main() {
	redisAddr := flag.String("redis-addr", "redis:6379", "Redis server address")
	streamKey := flag.String("stream-key", "topology-stream", "Redis stream key for topology")
	snapshotKey := flag.String("snapshot-key", "topology-snapshot", "Redis key for current snapshot")
	numSatellites := flag.Int("num-sats", 500, "Number of satellites to simulate")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Initialize Redis publisher
	pub, err := publisher.NewRedisPublisher(*redisAddr, *streamKey, *snapshotKey)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer pub.Close()

	log.Println("✓ Connected to Redis")

	// Initialize orbital propagator
	propagator := kinematics.NewOrbitalPropagator()

	// Generate 500 satellites in constellation pattern
	// Typical LEO constellations: 72 orbital planes × 22 satellites per plane
	numPlanes := 72
	satsPerPlane := (*numSatellites) / numPlanes
	if satsPerPlane == 0 {
		satsPerPlane = 1
	}

	satID := 0
	leoAltitude := 550e3 // 550 km altitude (typical LEO)

	for plane := 0; plane < numPlanes; plane++ {
		raan := float64(plane) * (2 * math.Pi / float64(numPlanes))

		for slot := 0; slot < satsPerPlane; slot++ {
			meanAnom := float64(slot) * (2 * math.Pi / float64(satsPerPlane))

			elem := &kinematics.OrbitalElements{
				SemiMajorAxis: 6.371e6 + leoAltitude,
				Eccentricity:  0.0,
				Inclination:   math.Pi * 51.6 / 180, // 51.6° (Starlink-like)
				RAAN:          raan,
				ArgPerigee:    0,
				MeanAnomaly:   meanAnom,
			}

			propagator.AddSatellite(satID, elem)
			satID++
		}
	}

	log.Printf("✓ Initialized %d satellites", satID)

	// Initialize topology calculator
	los := topology.NewLineOfSight(propagator.GetEarthRadius())

	// Main simulation loop at 10 Hz
	ticker := time.NewTicker(time.Second / 10)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var loopCount int64
	log.Println("Orbital engine running, broadcasting topology at 10 Hz...")

	for {
		select {
		case <-sigChan:
			log.Println("Shutting down orbital engine...")
			return

		case <-ticker.C:
			loopCount++
			currentTime := time.Now()

			// Propagate all satellites
			states := propagator.PropagateSatellites(currentTime)

			// Convert to position map
			positionMap := make(map[int][3]float64)
			for _, state := range states {
				positionMap[state.SatelliteID] = state.Position
			}

			// Compute adjacency (line-of-sight topology)
			adjacency := los.ComputeAdjacencyMatrix(positionMap)

			// Build snapshot for publishing
			entries := make(map[int]publisher.AdjacencyEntry)
			for satID, neighbors := range adjacency {
				state := findState(states, satID)
				entries[satID] = publisher.AdjacencyEntry{
					Timestamp: currentTime,
					Satellite: satID,
					Adjacency: neighbors,
					Position:  state.Position,
					Velocity:  state.Velocity,
				}
			}

			snapshot := &publisher.StateSnapshot{
				Timestamp: currentTime,
				Entries:   entries,
			}

			// Publish to Redis
			if err := pub.PublishAdjacencyMatrix(snapshot); err != nil {
				log.Printf("Error publishing: %v", err)
			}

			if *verbose && loopCount%10 == 0 {
				connectedPairs := 0
				for _, neighbors := range adjacency {
					connectedPairs += len(neighbors)
				}
				connectedPairs /= 2 // Each link counted twice

				log.Printf("Loop %d: %d satellites, %d connected pairs", loopCount, len(adjacency), connectedPairs)
			}
		}
	}
}

// findState finds a satellite state by ID
func findState(states []kinematics.SatelliteState, satID int) kinematics.SatelliteState {
	for _, s := range states {
		if s.SatelliteID == satID {
			return s
		}
	}
	return kinematics.SatelliteState{}
}
