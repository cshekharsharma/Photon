package contract

import (
	"context"

	"github.com/cshekharsharma/photon/cloud/entity/visualanalysis"
)

// VisualAnalysisInterface is implemented by cloud providers that can analyse images.
type VisualAnalysisInterface interface {
	DetectFaces(ctx context.Context, input *visualanalysis.DetectFacesInput) (*visualanalysis.DetectFacesResult, error)
	DetectLabels(ctx context.Context, input *visualanalysis.DetectLabelsInput) (*visualanalysis.DetectLabelsResult, error)
	CompareFaces(ctx context.Context, input *visualanalysis.CompareFacesInput) (*visualanalysis.CompareFacesResult, error)
}
