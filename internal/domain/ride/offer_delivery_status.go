package ride

type OfferDeliveryStatus string

const (
	OfferDeliveryOffered       OfferDeliveryStatus = "OFFERED"
	OfferDeliveryWebSocketSent OfferDeliveryStatus = "DELIVERED_WS"
	OfferDeliveryPushSent      OfferDeliveryStatus = "DELIVERED_PUSH"
	OfferDeliveryAcked         OfferDeliveryStatus = "ACKED"
	OfferDeliveryPushFailed    OfferDeliveryStatus = "PUSH_FAILED"
	OfferDeliveryWsFailed      OfferDeliveryStatus = "WS_FAILED"
)
