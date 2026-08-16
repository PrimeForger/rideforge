package ranking_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/ranking"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/routing"
	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/google/uuid"
)

func defaultTestRankingConfig() *config.RankingConfig {
	return &config.RankingConfig{
		DistanceWeight:     0.35,
		AcceptanceWeight:   0.20,
		CancellationWeight: 0.15,
		TimeoutWeight:      0.10,
		RatingWeight:       0.15,
		ExperienceWeight:   0.10,
		FairnessWeight:     0.05,
		MaxETASeconds:      1200.0, // 20 minutes
	}
}

// zeroQualityFeatures returns QualityFeatures that produce 0.0 quality score in QualityCalculator
func zeroQualityFeatures() ranking.QualityFeatures {
	return ranking.QualityFeatures{
		CancellationRate: 1.0, // 1 - 1 = 0 penalty
		TimeoutRate:      1.0, // 1 - 1 = 0 penalty
	}
}

type fakeRouteProvider struct {
	result routing.RouteResult
	err    error
}

func (p *fakeRouteProvider) CalculateRoute(ctx context.Context, req routing.RouteRequest) (routing.RouteResult, error) {
	if ctx.Err() != nil {
		return routing.RouteResult{}, ctx.Err()
	}
	return p.result, p.err
}

func TestDefaultScorer_ETAAvailable_PrimaryTravelSignal(t *testing.T) {
	cfg := defaultTestRankingConfig()
	scorer := ranking.NewDefaultScorer(cfg)
	ctx := context.Background()

	// Candidate with ETA 240s (4 min) and Distance 10km
	features1 := ranking.Features{
		Travel: ranking.TravelFeatures{
			DistanceKm: 10.0,
			ETASeconds: 240.0,
		},
		Quality: zeroQualityFeatures(),
	}

	// Candidate with same ETA 240s (4 min) but Distance 2km
	features2 := ranking.Features{
		Travel: ranking.TravelFeatures{
			DistanceKm: 2.0,
			ETASeconds: 240.0,
		},
		Quality: zeroQualityFeatures(),
	}

	score1, err := scorer.Score(ctx, features1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	score2, err := scorer.Score(ctx, features2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Both have same ETA, so travel score must be identical regardless of distance
	if math.Abs(score1.Value-score2.Value) > 0.00001 {
		t.Errorf("expected identical travel score when ETA is available, got %f vs %f", score1.Value, score2.Value)
	}

	// Expected ETA score = 1 - (240 / 1200) = 0.8
	// Travel contribution = 0.35 * 0.8 = 0.28
	expectedScore := 0.35 * 0.8
	if math.Abs(score1.Value-expectedScore) > 0.00001 {
		t.Errorf("expected score %f, got %f", expectedScore, score1.Value)
	}
}

func TestDefaultScorer_ETANormalization_Boundaries(t *testing.T) {
	cfg := defaultTestRankingConfig()
	scorer := ranking.NewDefaultScorer(cfg)
	ctx := context.Background()

	// Low ETA (120s / 2 min) -> score = 1 - (120/1200) = 0.90 -> travel contrib = 0.35 * 0.90 = 0.315
	fLow := ranking.Features{
		Travel:  ranking.TravelFeatures{DistanceKm: 5.0, ETASeconds: 120.0},
		Quality: zeroQualityFeatures(),
	}
	sLow, _ := scorer.Score(ctx, fLow)
	expectedLow := 0.35 * 0.90
	if math.Abs(sLow.Value-expectedLow) > 0.00001 {
		t.Errorf("low ETA score mismatch: expected %f, got %f", expectedLow, sLow.Value)
	}

	// High ETA (1080s / 18 min) -> score = 1 - (1080/1200) = 0.10 -> travel contrib = 0.35 * 0.10 = 0.035
	fHigh := ranking.Features{
		Travel:  ranking.TravelFeatures{DistanceKm: 5.0, ETASeconds: 1080.0},
		Quality: zeroQualityFeatures(),
	}
	sHigh, _ := scorer.Score(ctx, fHigh)
	expectedHigh := 0.35 * 0.10
	if math.Abs(sHigh.Value-expectedHigh) > 0.00001 {
		t.Errorf("high ETA score mismatch: expected %f, got %f", expectedHigh, sHigh.Value)
	}

	// ETA exceeding MaxETASeconds (1500s > 1200s) -> score = 0
	fExceed := ranking.Features{
		Travel:  ranking.TravelFeatures{DistanceKm: 5.0, ETASeconds: 1500.0},
		Quality: zeroQualityFeatures(),
	}
	sExceed, _ := scorer.Score(ctx, fExceed)
	if sExceed.Value != 0.0 {
		t.Errorf("expected score 0 for ETA >= MaxETASeconds, got %f", sExceed.Value)
	}
}

func TestDefaultScorer_ETAUnavailable_DistanceFallback(t *testing.T) {
	cfg := defaultTestRankingConfig()
	scorer := ranking.NewDefaultScorer(cfg)
	ctx := context.Background()

	// ETASeconds = 0 (unavailable), DistanceKm = 4.0
	// Expected distance score = 1 / (1 + 4.0) = 0.20
	// Travel contribution = 0.35 * 0.20 = 0.07
	f := ranking.Features{
		Travel: ranking.TravelFeatures{
			DistanceKm: 4.0,
			ETASeconds: 0,
		},
		Quality: zeroQualityFeatures(),
	}

	score, err := scorer.Score(ctx, f)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedScore := 0.35 * (1.0 / 5.0)
	if math.Abs(score.Value-expectedScore) > 0.00001 {
		t.Errorf("expected distance fallback score %f, got %f", expectedScore, score.Value)
	}
}

func TestDefaultScorer_CandidateOrdering_ETABased(t *testing.T) {
	cfg := defaultTestRankingConfig()
	ctx := context.Background()
	pCtx := &pipeline.Context{PickupLat: 12.9716, PickupLng: 77.5946}

	// Driver A: ETA 4 min (240s)
	rankerA := ranking.NewDefaultRanker(
		ranking.NewDefaultFeatureExtractor(ranking.NewDefaultTravelCalculator(&fakeRouteProvider{
			result: routing.RouteResult{DistanceKm: 5.0, ETASeconds: 240.0},
		})),
		ranking.NewDefaultScorer(cfg),
	)

	// Driver B: ETA 8 min (480s)
	rankerB := ranking.NewDefaultRanker(
		ranking.NewDefaultFeatureExtractor(ranking.NewDefaultTravelCalculator(&fakeRouteProvider{
			result: routing.RouteResult{DistanceKm: 5.0, ETASeconds: 480.0},
		})),
		ranking.NewDefaultScorer(cfg),
	)

	// Driver C: ETA 12 min (720s)
	rankerC := ranking.NewDefaultRanker(
		ranking.NewDefaultFeatureExtractor(ranking.NewDefaultTravelCalculator(&fakeRouteProvider{
			result: routing.RouteResult{DistanceKm: 5.0, ETASeconds: 720.0},
		})),
		ranking.NewDefaultScorer(cfg),
	)

	c := &candidate.Candidate{
		ID: uuid.New(),
		Driver: &driver.Driver{
			ID:               uuid.New(),
			Rating:           4.8,
			AcceptanceRate:   1.0,
			CancellationRate: 1.0,
			TimeoutRate:      1.0,
		},
	}

	scoreA, _ := rankerA.Rank(ctx, pCtx, c)
	scoreB, _ := rankerB.Rank(ctx, pCtx, c)
	scoreC, _ := rankerC.Rank(ctx, pCtx, c)

	if !(scoreA.Value > scoreB.Value && scoreB.Value > scoreC.Value) {
		t.Errorf("expected candidate ordering A > B > C based on ETA, got A=%f, B=%f, C=%f",
			scoreA.Value, scoreB.Value, scoreC.Value)
	}
}

func TestDefaultScorer_QualityAndFairnessPreservation(t *testing.T) {
	cfg := defaultTestRankingConfig()
	scorer := ranking.NewDefaultScorer(cfg)
	ctx := context.Background()

	now := time.Now()

	// High quality driver
	fHighQuality := ranking.Features{
		Travel: ranking.TravelFeatures{DistanceKm: 2.0, ETASeconds: 300.0},
		Quality: ranking.QualityFeatures{
			AcceptanceRate:   0.98,
			CancellationRate: 0.01,
			TimeoutRate:      0.01,
			Rating:           4.9,
			CompletedTrips:   500,
		},
		Fairness: ranking.FairnessFeatures{LastAssignedAt: now.Add(-30 * time.Minute)},
	}

	// Low quality driver (same ETA/distance and fairness)
	fLowQuality := ranking.Features{
		Travel: ranking.TravelFeatures{DistanceKm: 2.0, ETASeconds: 300.0},
		Quality: ranking.QualityFeatures{
			AcceptanceRate:   0.60,
			CancellationRate: 0.20,
			TimeoutRate:      0.15,
			Rating:           3.5,
			CompletedTrips:   10,
		},
		Fairness: ranking.FairnessFeatures{LastAssignedAt: now.Add(-30 * time.Minute)},
	}

	sHQ, _ := scorer.Score(ctx, fHighQuality)
	sLQ, _ := scorer.Score(ctx, fLowQuality)

	if sHQ.Value <= sLQ.Value {
		t.Errorf("expected high quality driver score (%f) to be strictly greater than low quality driver score (%f)",
			sHQ.Value, sLQ.Value)
	}
}

func TestDefaultScorer_EdgeCases(t *testing.T) {
	cfg := defaultTestRankingConfig()
	scorer := ranking.NewDefaultScorer(cfg)
	ctx := context.Background()

	// 1. Negative ETASeconds -> falls back to distance score
	fNegETA := ranking.Features{
		Travel:  ranking.TravelFeatures{DistanceKm: 3.0, ETASeconds: -100.0},
		Quality: zeroQualityFeatures(),
	}
	sNegETA, err := scorer.Score(ctx, fNegETA)
	if err != nil {
		t.Fatalf("expected no error on negative ETA, got %v", err)
	}
	expectedDistFallback := 0.35 * (1.0 / (1.0 + 3.0))
	if math.Abs(sNegETA.Value-expectedDistFallback) > 0.00001 {
		t.Errorf("expected distance fallback score %f for negative ETA, got %f", expectedDistFallback, sNegETA.Value)
	}

	// 2. Negative DistanceKm -> safe 0.0 distance score
	fNegDist := ranking.Features{
		Travel:  ranking.TravelFeatures{DistanceKm: -5.0, ETASeconds: 0},
		Quality: zeroQualityFeatures(),
	}
	sNegDist, err := scorer.Score(ctx, fNegDist)
	if err != nil {
		t.Fatalf("expected no error on negative distance, got %v", err)
	}
	if sNegDist.Value != 0.0 {
		t.Errorf("expected 0.0 score for negative distance, got %f", sNegDist.Value)
	}

	// 3. MaxETASeconds <= 0 -> safe 0.0 ETA score
	cfgZeroMax := defaultTestRankingConfig()
	cfgZeroMax.MaxETASeconds = 0.0
	scorerZeroMax := ranking.NewDefaultScorer(cfgZeroMax)
	fZeroMax := ranking.Features{
		Travel:  ranking.TravelFeatures{DistanceKm: 2.0, ETASeconds: 300.0},
		Quality: zeroQualityFeatures(),
	}
	sZeroMax, err := scorerZeroMax.Score(ctx, fZeroMax)
	if err != nil {
		t.Fatalf("expected no error on zero MaxETASeconds, got %v", err)
	}
	if sZeroMax.Value != 0.0 {
		t.Errorf("expected 0.0 score when MaxETASeconds <= 0, got %f", sZeroMax.Value)
	}
}

func TestDefaultScorer_FullRoutingFallbackIntegration(t *testing.T) {
	// Full integration test:
	// Routing provider disabled -> Haversine fallback -> ETASeconds = 0 -> Distance fallback score
	cfg := defaultTestRankingConfig()
	disabledRouting := routing.NewDisabledProvider()
	travelCalc := ranking.NewDefaultTravelCalculator(disabledRouting)
	extractor := ranking.NewDefaultFeatureExtractor(travelCalc)
	ranker := ranking.NewDefaultRanker(extractor, ranking.NewDefaultScorer(cfg))

	ctx := context.Background()
	pCtx := &pipeline.Context{
		PickupLat: 12.9716,
		PickupLng: 77.5946,
	}

	d := &driver.Driver{
		ID:  uuid.New(),
		Lat: 12.9750,
		Lng: 77.5990,
	}
	c := &candidate.Candidate{
		ID:     uuid.New(),
		Driver: d,
	}

	score, err := ranker.Rank(ctx, pCtx, c)
	if err != nil {
		t.Fatalf("expected no error on full routing fallback chain, got %v", err)
	}

	if score.Value <= 0 {
		t.Errorf("expected positive score on full fallback chain, got %f", score.Value)
	}
}
