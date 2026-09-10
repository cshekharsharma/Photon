package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	rekognitiontypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/cshekharsharma/photon/cloud/entity/visualanalysis"
)

// awsRekognitionClientInterface abstracts AWS Rekognition for testing.
type awsRekognitionClientInterface interface {
	DetectFaces(ctx context.Context, params *rekognition.DetectFacesInput, optFns ...func(*rekognition.Options)) (*rekognition.DetectFacesOutput, error)
	DetectLabels(ctx context.Context, params *rekognition.DetectLabelsInput, optFns ...func(*rekognition.Options)) (*rekognition.DetectLabelsOutput, error)
	CompareFaces(ctx context.Context, params *rekognition.CompareFacesInput, optFns ...func(*rekognition.Options)) (*rekognition.CompareFacesOutput, error)
}

// VisualAnalysis provides image-analysis methods backed by AWS Rekognition.
type VisualAnalysis struct {
	RekognitionClient awsRekognitionClientInterface
}

// DetectFaces runs face detection for an object-storage image.
func (va VisualAnalysis) DetectFaces(ctx context.Context, input *visualanalysis.DetectFacesInput) (*visualanalysis.DetectFacesResult, error) {
	attrs := []rekognitiontypes.Attribute{rekognitiontypes.AttributeDefault}
	if input != nil && len(input.Attributes) > 0 {
		attrs = make([]rekognitiontypes.Attribute, 0, len(input.Attributes))
		for _, attr := range input.Attributes {
			attrs = append(attrs, rekognitiontypes.Attribute(attr))
		}
	}
	awsResult, err := va.RekognitionClient.DetectFaces(ctx, &rekognition.DetectFacesInput{
		Image:      toRekognitionImage(detectFacesImage(input)),
		Attributes: attrs,
	})
	if err != nil {
		return nil, err
	}
	result := &visualanalysis.DetectFacesResult{
		FaceCount: len(awsResult.FaceDetails),
		Faces:     make([]visualanalysis.FaceDetail, 0, len(awsResult.FaceDetails)),
	}
	for _, face := range awsResult.FaceDetails {
		result.Faces = append(result.Faces, toFaceDetail(face))
	}
	return result, nil
}

// DetectLabels runs label detection for an object-storage image.
func (va VisualAnalysis) DetectLabels(ctx context.Context, input *visualanalysis.DetectLabelsInput) (*visualanalysis.DetectLabelsResult, error) {
	minConfidence := float32(0)
	maxLabels := int32(100)
	if input != nil {
		minConfidence = input.MinConfidence
		if input.MaxLabels > 0 {
			maxLabels = input.MaxLabels
		}
	}
	awsResult, err := va.RekognitionClient.DetectLabels(ctx, &rekognition.DetectLabelsInput{
		Image:         toRekognitionImage(detectLabelsImage(input)),
		MaxLabels:     aws.Int32(maxLabels),
		MinConfidence: aws.Float32(minConfidence),
	})
	if err != nil {
		return nil, err
	}
	result := &visualanalysis.DetectLabelsResult{Labels: make([]visualanalysis.Label, 0, len(awsResult.Labels))}
	for _, label := range awsResult.Labels {
		parents := make([]string, 0, len(label.Parents))
		for _, parent := range label.Parents {
			parents = append(parents, aws.ToString(parent.Name))
		}
		result.Labels = append(result.Labels, visualanalysis.Label{
			Name:       aws.ToString(label.Name),
			Confidence: aws.ToFloat32(label.Confidence),
			Parents:    parents,
		})
	}
	return result, nil
}

// CompareFaces compares a source image with a target image.
func (va VisualAnalysis) CompareFaces(ctx context.Context, input *visualanalysis.CompareFacesInput) (*visualanalysis.CompareFacesResult, error) {
	threshold := float32(0)
	if input != nil {
		threshold = input.SimilarityThreshold
	}
	awsResult, err := va.RekognitionClient.CompareFaces(ctx, &rekognition.CompareFacesInput{
		SourceImage:         toRekognitionImage(sourceImage(input)),
		TargetImage:         toRekognitionImage(targetImage(input)),
		SimilarityThreshold: aws.Float32(threshold),
	})
	if err != nil {
		return nil, err
	}
	result := &visualanalysis.CompareFacesResult{
		MatchedFaces:   make([]visualanalysis.ComparedFace, 0, len(awsResult.FaceMatches)),
		UnmatchedCount: len(awsResult.UnmatchedFaces),
	}
	for _, match := range awsResult.FaceMatches {
		compared := visualanalysis.ComparedFace{Similarity: aws.ToFloat32(match.Similarity)}
		if match.Face != nil {
			compared.Face = toComparedFace(*match.Face)
		}
		if compared.Similarity > result.HighestSimilarity {
			result.HighestSimilarity = compared.Similarity
		}
		result.MatchedFaces = append(result.MatchedFaces, compared)
	}
	return result, nil
}

func detectFacesImage(input *visualanalysis.DetectFacesInput) visualanalysis.ImageObjectRef {
	if input == nil {
		return visualanalysis.ImageObjectRef{}
	}
	return input.Image
}

func detectLabelsImage(input *visualanalysis.DetectLabelsInput) visualanalysis.ImageObjectRef {
	if input == nil {
		return visualanalysis.ImageObjectRef{}
	}
	return input.Image
}

func sourceImage(input *visualanalysis.CompareFacesInput) visualanalysis.ImageObjectRef {
	if input == nil {
		return visualanalysis.ImageObjectRef{}
	}
	return input.SourceImage
}

func targetImage(input *visualanalysis.CompareFacesInput) visualanalysis.ImageObjectRef {
	if input == nil {
		return visualanalysis.ImageObjectRef{}
	}
	return input.TargetImage
}

func toRekognitionImage(ref visualanalysis.ImageObjectRef) *rekognitiontypes.Image {
	return &rekognitiontypes.Image{S3Object: &rekognitiontypes.S3Object{
		Bucket: aws.String(ref.Bucket),
		Name:   aws.String(ref.Key),
	}}
}

func toFaceDetail(face rekognitiontypes.FaceDetail) visualanalysis.FaceDetail {
	detail := visualanalysis.FaceDetail{
		Confidence: aws.ToFloat32(face.Confidence),
	}
	if face.BoundingBox != nil {
		detail.BoundingBox = visualanalysis.BoundingBox{
			Width:  aws.ToFloat32(face.BoundingBox.Width),
			Height: aws.ToFloat32(face.BoundingBox.Height),
			Left:   aws.ToFloat32(face.BoundingBox.Left),
			Top:    aws.ToFloat32(face.BoundingBox.Top),
		}
	}
	if face.Quality != nil {
		detail.Quality = visualanalysis.FaceQuality{
			Brightness: aws.ToFloat32(face.Quality.Brightness),
			Sharpness:  aws.ToFloat32(face.Quality.Sharpness),
		}
	}
	return detail
}

func toComparedFace(face rekognitiontypes.ComparedFace) visualanalysis.FaceDetail {
	detail := visualanalysis.FaceDetail{Confidence: aws.ToFloat32(face.Confidence)}
	if face.BoundingBox != nil {
		detail.BoundingBox = visualanalysis.BoundingBox{
			Width:  aws.ToFloat32(face.BoundingBox.Width),
			Height: aws.ToFloat32(face.BoundingBox.Height),
			Left:   aws.ToFloat32(face.BoundingBox.Left),
			Top:    aws.ToFloat32(face.BoundingBox.Top),
		}
	}
	return detail
}
