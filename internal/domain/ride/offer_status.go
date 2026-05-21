package ride

type OfferStatus string

const (
	OfferStatusOffered  OfferStatus = "OFFERED"
	OfferStatusAccepted OfferStatus = "ACCEPTED"
	OfferStatusRejected OfferStatus = "REJECTED"
	OfferStatusTimeout  OfferStatus = "TIMEOUT"
)

func (s OfferStatus) IsFinal() bool {
	return s == OfferStatusAccepted ||
		s == OfferStatusRejected ||
		s == OfferStatusTimeout
}

func (s OfferStatus) CanTransitionTo(next OfferStatus) bool {
	switch s {
	case OfferStatusOffered:
		return next == OfferStatusAccepted ||
			next == OfferStatusRejected ||
			next == OfferStatusTimeout
	default:
		return false
	}
}
