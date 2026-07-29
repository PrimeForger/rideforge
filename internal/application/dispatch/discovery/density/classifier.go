package density

type DensityClass int

const (
	DensityUnknown DensityClass = iota
	DensitySparse
	DensityNormal
	DensityDense
)

type DensityClassifier struct {
	sparseUpperBound int
	denseUpperBound  int
}

func NewDensityClassifier(
	sparseUpperBound int,
	denseUpperBound int,
) *DensityClassifier {

	return &DensityClassifier{
		sparseUpperBound: sparseUpperBound,
		denseUpperBound:  denseUpperBound,
	}
}

func (c *DensityClassifier) Classify(
	driverCount int,
) DensityClass {

	switch {

	case driverCount < c.sparseUpperBound:
		return DensitySparse

	case driverCount < c.denseUpperBound:
		return DensityNormal

	default:
		return DensityDense
	}
}

func (d DensityClass) String() string {
	switch d {
	case DensitySparse:
		return "sparse"
	case DensityNormal:
		return "normal"
	case DensityDense:
		return "dense"
	default:
		return "unknown"
	}
}
