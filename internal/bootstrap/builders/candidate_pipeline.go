package builders

import (
	candidatepipeline "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/ranking"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/stage"
)

func BuildCandidatePipeline(
	batchSource stage.BatchDataSource,
	offerProvider stage.OfferedDriverProvider,
	ranker ranking.Ranker,
) candidatepipeline.Pipeline {

	return candidatepipeline.New(
		stage.NewDefaultDriverLoader(batchSource),
		stage.NewValidationFilter(),
		stage.NewDuplicateFilter(),
		stage.NewAvailabilityFilter(),
		stage.NewAlreadyOfferedFilter(offerProvider),
		stage.NewRankingStage(ranker),
		stage.NewHeapBuilder(),
	)
}
