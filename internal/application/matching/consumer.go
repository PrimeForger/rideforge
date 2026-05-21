package matching

// import (
// 	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/kafka"
// )

// type MatchingService struct {
// 	producer *kafka.Producer
// }

// func NewMatchingService(producer *kafka.Producer) *MatchingService {
// 	return &MatchingService{producer: producer}
// }

// func (m *MatchingService) HandleRideRequested(ctx context.Context, message []byte) error {
// 	var event events.Envelope
// 	if err := json.Unmarshal(message, &event); err != nil {
// 		return err
// 	}

// 	if event.Type != ride.EventRideRequested {
// 		return nil
// 	}

// 	dataBytes, err := json.Marshal(event.Data)
// 	if err != nil {
// 		return err
// 	}

// 	var payload ride.RideRequested
// 	if err := json.Unmarshal(dataBytes, &payload); err != nil {
// 		return err
// 	}

// 	log.Println("Matching ride:", payload.RideID)

// 	// Simulate driver selection
// 	driverID := "driver-123"

// 	matchEvent := events.Envelope{
// 		ID:        uuid.NewString(),
// 		Type:      ride.EventDriverMatched,
// 		Aggregate: payload.RideID,
// 		Data: ride.DriverMatched{
// 			RideID:    payload.RideID,
// 			DriverID:  driverID,
// 			MatchedAt: time.Now(),
// 		},
// 		Occurred: time.Now(),
// 	}

// 	return m.producer.Publish(ctx, matchEvent)
// }
