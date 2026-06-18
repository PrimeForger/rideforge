package config

import (
	"os"
	"strconv"
	"time"
)

type RankingConfig struct {
	DistanceWeight     float64
	AcceptanceWeight   float64
	CancellationWeight float64
	TimeoutWeight      float64
	RatingWeight       float64
	ExperienceWeight   float64
	FairnessWeight     float64
}

type RealtimeConfig struct {
	MaxLocationAccuracyMeters float64
	MinLocationIntervalMs     int
	LocationSeqTTLSeconds     int
	HeartbeatTTLSeconds       int
	ConnectionTTLSeconds      int
	DisconnectTTLSeconds      int
	OfferDeliveryTTLSeconds   int
}

type Config struct {
	OfferBatchSize     int
	MaxDriverAttempts  int
	SearchRadiusKm     float64
	DriverOfferTimeout time.Duration

	Ranking  RankingConfig
	Realtime RealtimeConfig
}

func Load() *Config {
	cfg := &Config{
		OfferBatchSize:     getInt("OFFER_BATCH_SIZE", 3),
		MaxDriverAttempts:  getInt("MAX_DRIVER_ATTEMPTS", 5),
		SearchRadiusKm:     getFloat("SEARCH_RADIUS_KM", 5.0),
		DriverOfferTimeout: getDuration("DRIVER_OFFER_TIMEOUT", 10*time.Second),
		Ranking: RankingConfig{
			DistanceWeight:     getFloat("DISTANCE_WEIGHT", 0.35),
			AcceptanceWeight:   getFloat("ACCEPTANCE_WEIGHT", 0.20),
			CancellationWeight: getFloat("CANCELLATION_WEIGHT", 0.15),
			TimeoutWeight:      getFloat("TIMEOUT_WEIGHT", 0.10),
			RatingWeight:       getFloat("RATING_WEIGHT", 0.15),
			ExperienceWeight:   getFloat("EXPERIENCE_WEIGHT", 0.10),
			FairnessWeight:     getFloat("FAIRNESS_WEIGHT", 0.05),
		},
		Realtime: RealtimeConfig{
			MaxLocationAccuracyMeters: getFloat("MAX_LOCATION_ACCURACY_METERS", 100),
			MinLocationIntervalMs:     getInt("MIN_LOCATION_INTERVAL_MS", 1000),
			LocationSeqTTLSeconds:     getInt("LOCATION_SEQ_TTL_SECONDS", 86400),
			HeartbeatTTLSeconds:       getInt("DRIVER_HEARTBEAT_TTL_SECONDS", 30),
			ConnectionTTLSeconds:      getInt("DRIVER_CONNECTION_TTL_SECONDS", 60),
			DisconnectTTLSeconds:      getInt("DRIVER_DISCONNECT_TTL_SECONDS", 10),
			OfferDeliveryTTLSeconds:   getInt("OFFER_DELIVERY_TTL_SECONDS", 1800),
		},
	}
	cfg.normalizeWeights()
	return cfg
}

func (c *Config) normalizeWeights() {
	total :=
		c.Ranking.DistanceWeight +
			c.Ranking.AcceptanceWeight +
			c.Ranking.CancellationWeight +
			c.Ranking.TimeoutWeight +
			c.Ranking.RatingWeight +
			c.Ranking.ExperienceWeight +
			c.Ranking.FairnessWeight

	if total == 0 {
		return
	}

	c.Ranking.DistanceWeight /= total
	c.Ranking.AcceptanceWeight /= total
	c.Ranking.CancellationWeight /= total
	c.Ranking.TimeoutWeight /= total
	c.Ranking.RatingWeight /= total
	c.Ranking.ExperienceWeight /= total
	c.Ranking.FairnessWeight /= total
}

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func getDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
