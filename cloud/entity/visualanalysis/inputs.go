// Package visualanalysis provides cloud-agnostic inputs for image analysis.
package visualanalysis

// ImageObjectRef identifies an image already stored in object storage.
type ImageObjectRef struct {
	Bucket string
	Key    string
}

// DetectFacesInput describes an image face-detection request.
type DetectFacesInput struct {
	Image      ImageObjectRef
	Attributes []string
}

// DetectLabelsInput describes an image label-detection request.
type DetectLabelsInput struct {
	Image         ImageObjectRef
	MaxLabels     int32
	MinConfidence float32
}

// CompareFacesInput describes a source-vs-target face comparison request.
type CompareFacesInput struct {
	SourceImage         ImageObjectRef
	TargetImage         ImageObjectRef
	SimilarityThreshold float32
}
