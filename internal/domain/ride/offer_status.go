package ride

type OfferStatus string

const (
	OfferStatusOffered  OfferStatus = "OFFERED"
	OfferStatusAccepted OfferStatus = "ACCEPTED"
	OfferStatusRejected OfferStatus = "REJECTED"
	OfferStatusTimeout  OfferStatus = "TIMEOUT"
	OfferStatusExpired  OfferStatus = "EXPIRED"
)

func (s OfferStatus) IsFinal() bool {
	return s == OfferStatusAccepted ||
		s == OfferStatusRejected ||
		s == OfferStatusTimeout ||
		s == OfferStatusExpired
}

func (s OfferStatus) CanTransitionTo(next OfferStatus) bool {
	switch s {
	case OfferStatusOffered:
		return next == OfferStatusAccepted ||
			next == OfferStatusRejected ||
			next == OfferStatusTimeout ||
			next == OfferStatusExpired
	default:
		return false
	}
}
