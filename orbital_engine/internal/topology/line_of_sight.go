package topology

import (
	"math"
)

// LineOfSight detector using raycasting to determine if Earth blocks two satellites
type LineOfSight struct {
	earthRadius float64
}

// NewLineOfSight creates a new line-of-sight calculator
func NewLineOfSight(earthRadius float64) *LineOfSight {
	return &LineOfSight{earthRadius: earthRadius}
}

// CanSee determines if satellite A can see satellite B (no Earth obstruction)
// Uses raycasting: if the line segment AB passes within earthRadius of Earth center, blocked
func (los *LineOfSight) CanSee(posA, posB [3]float64) bool {
	// Vector from A to B
	dx := posB[0] - posA[0]
	dy := posB[1] - posA[1]
	dz := posB[2] - posA[2]

	// Distance from Earth center to line AB is the key metric
	// Parametric form: P(t) = A + t(B-A) for t in [0,1]
	// Closest point on line to Earth center (0,0,0)

	// t_closest = -A·D / (D·D) where D = B-A
	dotAD := posA[0]*dx + posA[1]*dy + posA[2]*dz
	dotDD := dx*dx + dy*dy + dz*dz

	if dotDD == 0 {
		// Satellites at same position
		return true
	}

	t := -dotAD / dotDD

	// Clamp t to [0, 1] for line segment (not infinite line)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	// Closest point on segment
	closestX := posA[0] + t*dx
	closestY := posA[1] + t*dy
	closestZ := posA[2] + t*dz

	// Distance from Earth center to closest point
	distToEarth := math.Sqrt(closestX*closestX + closestY*closestY + closestZ*closestZ)

	// If distance >= earthRadius, line is not blocked
	return distToEarth >= los.earthRadius
}

// ComputeAdjacencyMatrix builds the connectivity graph for all satellites
// Returns a map: from satID -> []toSatIDs (can see each other)
func (los *LineOfSight) ComputeAdjacencyMatrix(states map[int][3]float64) map[int][]int {
	adjacency := make(map[int][]int)

	satIDs := make([]int, 0, len(states))
	for id := range states {
		satIDs = append(satIDs, id)
		adjacency[id] = []int{}
	}

	// Check all pairs
	for i, idA := range satIDs {
		for j, idB := range satIDs {
			if i < j { // Avoid checking twice
				if los.CanSee(states[idA], states[idB]) {
					adjacency[idA] = append(adjacency[idA], idB)
					adjacency[idB] = append(adjacency[idB], idA)
				}
			}
		}
	}

	return adjacency
}

// ComputeDistance calculates Euclidean distance between two positions
// Used for weighting edges in the routing graph
func ComputeDistance(pos1, pos2 [3]float64) float64 {
	dx := pos2[0] - pos1[0]
	dy := pos2[1] - pos1[1]
	dz := pos2[2] - pos1[2]
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// EstimateLatency estimates transmission latency based on distance
// Approximate model: latency = distance / speed_of_light + overhead
func EstimateLatency(distance float64) float64 {
	speedOfLight := 3e8 // m/s
	overhead := 0.001   // 1ms fixed overhead
	return distance/speedOfLight + overhead
}
