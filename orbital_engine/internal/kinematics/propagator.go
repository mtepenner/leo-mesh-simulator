package kinematics

import (
	"math"
	"time"
)

// OrbitalElements represents TLE-derived orbital parameters
type OrbitalElements struct {
	SemiMajorAxis float64 // meters
	Eccentricity  float64
	Inclination   float64 // radians
	RAAN          float64 // Right Ascension of Ascending Node (radians)
	ArgPerigee    float64 // Argument of Perigee (radians)
	MeanAnomaly   float64 // Mean Anomaly (radians)
}

// SatelliteState represents position and velocity in ECEF frame
type SatelliteState struct {
	SatelliteID int
	Timestamp   time.Time
	Position    [3]float64 // X, Y, Z in meters
	Velocity    [3]float64 // Vx, Vy, Vz in m/s
}

// OrbitalPropagator handles SGP4-style propagation using simplified Kepler dynamics
type OrbitalPropagator struct {
	satellites  map[int]*OrbitalElements
	mu          float64 // Earth's gravitational parameter (m^3/s^2)
	earthRadius float64 // Earth radius in meters
}

// NewOrbitalPropagator creates a new propagator with Earth parameters
func NewOrbitalPropagator() *OrbitalPropagator {
	return &OrbitalPropagator{
		satellites:  make(map[int]*OrbitalElements),
		mu:          3.986004418e14, // Earth GM (m^3/s^2)
		earthRadius: 6.371e6,        // Earth radius (m)
	}
}

// AddSatellite registers a satellite with TLE-derived orbital elements
func (op *OrbitalPropagator) AddSatellite(id int, elements *OrbitalElements) {
	op.satellites[id] = elements
}

// PropagateSatellites computes current positions of all satellites at time t
func (op *OrbitalPropagator) PropagateSatellites(currentTime time.Time) []SatelliteState {
	var states []SatelliteState

	for satID, elem := range op.satellites {
		// Update mean anomaly based on elapsed time (simplified: assume constant mean motion)
		// For LEO (500km altitude), orbital period ≈ 5600 seconds
		a := elem.SemiMajorAxis
		n := math.Sqrt(op.mu / (a * a * a)) // Mean motion (rad/s)

		// Compute new mean anomaly (with day wrapping for numerical stability)
		timeElapsed := currentTime.Sub(time.Now()).Seconds()
		newMeanAnomaly := elem.MeanAnomaly + n*timeElapsed

		// Solve Kepler's equation: E - e*sin(E) = M (using Newton-Raphson)
		E := op.solveKeplerEquation(newMeanAnomaly, elem.Eccentricity)

		// Compute true anomaly
		trueAnom := 2.0 * math.Atan2(
			math.Sqrt(1+elem.Eccentricity)*math.Sin(E/2.0),
			math.Sqrt(1-elem.Eccentricity)*math.Cos(E/2.0),
		)

		// Compute position in orbital plane (perifocal frame)
		r := a * (1 - elem.Eccentricity*elem.Eccentricity) / (1 + elem.Eccentricity*math.Cos(trueAnom))
		posOrbital := [3]float64{
			r * math.Cos(trueAnom),
			r * math.Sin(trueAnom),
			0,
		}

		// Compute velocity in orbital plane
		h := math.Sqrt(op.mu * a * (1 - elem.Eccentricity*elem.Eccentricity))
		velOrbital := [3]float64{
			-op.mu / h * math.Sin(trueAnom),
			op.mu / h * (elem.Eccentricity + math.Cos(trueAnom)),
			0,
		}

		// Transform from perifocal to ECEF frame using Euler angles (RAAN, inclination, arg of perigee)
		pos := op.perifocalToECEF(posOrbital, elem)
		vel := op.perifocalToECEF(velOrbital, elem)

		states = append(states, SatelliteState{
			SatelliteID: satID,
			Timestamp:   currentTime,
			Position:    pos,
			Velocity:    vel,
		})
	}

	return states
}

// solveKeplerEquation solves E - e*sin(E) = M using Newton-Raphson
func (op *OrbitalPropagator) solveKeplerEquation(M, e float64) float64 {
	// Normalize mean anomaly to [0, 2π)
	M = math.Mod(M, 2*math.Pi)
	if M < 0 {
		M += 2 * math.Pi
	}

	// Initial guess for eccentric anomaly
	E := M
	if e > 0.8 {
		E = math.Pi
	}

	// Newton-Raphson iteration
	for i := 0; i < 10; i++ {
		f := E - e*math.Sin(E) - M
		fPrime := 1 - e*math.Cos(E)
		E = E - f/fPrime

		if math.Abs(f) < 1e-12 {
			break
		}
	}

	return E
}

// perifocalToECEF transforms coordinates from perifocal frame to ECEF
func (op *OrbitalPropagator) perifocalToECEF(perifocal [3]float64, elem *OrbitalElements) [3]float64 {
	// Rotation matrix from perifocal to ECEF (composed of three Euler rotations)
	// R = Rz(RAAN) * Rx(inclination) * Rz(arg of perigee)

	// Simplification: Use rotation matrices
	cosRAAN := math.Cos(elem.RAAN)
	sinRAAN := math.Sin(elem.RAAN)
	cosInc := math.Cos(elem.Inclination)
	sinInc := math.Sin(elem.Inclination)
	cosArg := math.Cos(elem.ArgPerigee)
	sinArg := math.Sin(elem.ArgPerigee)

	// Intermediate rotation: Perifocal to intermediate frame
	x1 := cosArg*perifocal[0] - sinArg*perifocal[1]
	y1 := sinArg*perifocal[0] + cosArg*perifocal[1]
	z1 := perifocal[2]

	// Further rotation
	x2 := x1
	y2 := cosInc*y1 - sinInc*z1
	z2 := sinInc*y1 + cosInc*z1

	// Final rotation (RAAN)
	x := cosRAAN*x2 - sinRAAN*y2
	y := sinRAAN*x2 + cosRAAN*y2
	z := z2

	return [3]float64{x, y, z}
}

// GetEarthRadius returns Earth's radius for line-of-sight calculations
func (op *OrbitalPropagator) GetEarthRadius() float64 {
	return op.earthRadius
}

// GetGravitationalParameter returns GM for orbital calculations
func (op *OrbitalPropagator) GetGravitationalParameter() float64 {
	return op.mu
}
