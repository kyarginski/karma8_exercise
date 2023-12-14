package health

type LivenessChecker interface {
	LivenessCheck() bool
}

type ReadinessChecker interface {
	ReadinessCheck() bool
}
