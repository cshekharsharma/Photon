// Package visualanalysis provides cloud-agnostic image analysis results.
package visualanalysis

// BoundingBox describes a normalized image region.
type BoundingBox struct {
	Width  float32
	Height float32
	Left   float32
	Top    float32
}

// FaceQuality captures provider face-quality signals.
type FaceQuality struct {
	Brightness float32
	Sharpness  float32
}

// FaceDetail describes one detected face.
type FaceDetail struct {
	Confidence  float32
	BoundingBox BoundingBox
	Quality     FaceQuality
}

// DetectFacesResult contains face-detection output.
type DetectFacesResult struct {
	FaceCount int
	Faces     []FaceDetail
}

// Label describes an object/scene label.
type Label struct {
	Name       string
	Confidence float32
	Parents    []string
}

// DetectLabelsResult contains label-detection output.
type DetectLabelsResult struct {
	Labels []Label
}

// ComparedFace describes one matched face in a comparison result.
type ComparedFace struct {
	Similarity float32
	Face       FaceDetail
}

// CompareFacesResult contains source-vs-target face comparison output.
type CompareFacesResult struct {
	MatchedFaces      []ComparedFace
	UnmatchedCount    int
	HighestSimilarity float32
}
