package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	rekognitiontypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/cshekharsharma/photon/cloud/entity/visualanalysis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRekognitionClient struct {
	mock.Mock
}

func (m *mockRekognitionClient) DetectFaces(ctx context.Context, params *rekognition.DetectFacesInput, optFns ...func(*rekognition.Options)) (*rekognition.DetectFacesOutput, error) {
	args := m.Called(ctx, params)
	if out, ok := args.Get(0).(*rekognition.DetectFacesOutput); ok {
		return out, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRekognitionClient) DetectLabels(ctx context.Context, params *rekognition.DetectLabelsInput, optFns ...func(*rekognition.Options)) (*rekognition.DetectLabelsOutput, error) {
	args := m.Called(ctx, params)
	if out, ok := args.Get(0).(*rekognition.DetectLabelsOutput); ok {
		return out, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRekognitionClient) CompareFaces(ctx context.Context, params *rekognition.CompareFacesInput, optFns ...func(*rekognition.Options)) (*rekognition.CompareFacesOutput, error) {
	args := m.Called(ctx, params)
	if out, ok := args.Get(0).(*rekognition.CompareFacesOutput); ok {
		return out, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestVisualAnalysisDetectFaces(t *testing.T) {
	client := new(mockRekognitionClient)
	service := VisualAnalysis{RekognitionClient: client}
	client.On("DetectFaces", mock.Anything, mock.MatchedBy(func(in *rekognition.DetectFacesInput) bool {
		return aws.ToString(in.Image.S3Object.Bucket) == "bucket" &&
			aws.ToString(in.Image.S3Object.Name) == "key" &&
			len(in.Attributes) == 1 &&
			string(in.Attributes[0]) == "ALL"
	})).Return(&rekognition.DetectFacesOutput{FaceDetails: []rekognitiontypes.FaceDetail{{
		Confidence: aws.Float32(99.1),
		BoundingBox: &rekognitiontypes.BoundingBox{
			Width: aws.Float32(0.2), Height: aws.Float32(0.3), Left: aws.Float32(0.4), Top: aws.Float32(0.5),
		},
		Quality: &rekognitiontypes.ImageQuality{Brightness: aws.Float32(88), Sharpness: aws.Float32(77)},
	}}}, nil)

	result, err := service.DetectFaces(context.Background(), &visualanalysis.DetectFacesInput{
		Image:      visualanalysis.ImageObjectRef{Bucket: "bucket", Key: "key"},
		Attributes: []string{"ALL"},
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, result.FaceCount)
	assert.Equal(t, float32(99.1), result.Faces[0].Confidence)
	assert.Equal(t, float32(0.4), result.Faces[0].BoundingBox.Left)
	assert.Equal(t, float32(77), result.Faces[0].Quality.Sharpness)
}

func TestVisualAnalysisDetectFacesError(t *testing.T) {
	client := new(mockRekognitionClient)
	service := VisualAnalysis{RekognitionClient: client}
	client.On("DetectFaces", mock.Anything, mock.Anything).Return((*rekognition.DetectFacesOutput)(nil), errors.New("faces failed"))
	_, err := service.DetectFaces(context.Background(), &visualanalysis.DetectFacesInput{})
	assert.Error(t, err)
}

func TestVisualAnalysisDetectLabels(t *testing.T) {
	client := new(mockRekognitionClient)
	service := VisualAnalysis{RekognitionClient: client}
	client.On("DetectLabels", mock.Anything, mock.MatchedBy(func(in *rekognition.DetectLabelsInput) bool {
		return aws.ToInt32(in.MaxLabels) == 3 && aws.ToFloat32(in.MinConfidence) == 70
	})).Return(&rekognition.DetectLabelsOutput{Labels: []rekognitiontypes.Label{{
		Name: aws.String("Cell Phone"), Confidence: aws.Float32(91), Parents: []rekognitiontypes.Parent{{Name: aws.String("Electronics")}},
	}}}, nil)

	result, err := service.DetectLabels(context.Background(), &visualanalysis.DetectLabelsInput{
		Image: visualanalysis.ImageObjectRef{Bucket: "bucket", Key: "key"}, MaxLabels: 3, MinConfidence: 70,
	})
	assert.NoError(t, err)
	assert.Equal(t, "Cell Phone", result.Labels[0].Name)
	assert.Equal(t, []string{"Electronics"}, result.Labels[0].Parents)
}

func TestVisualAnalysisDetectLabelsError(t *testing.T) {
	client := new(mockRekognitionClient)
	service := VisualAnalysis{RekognitionClient: client}
	client.On("DetectLabels", mock.Anything, mock.Anything).Return((*rekognition.DetectLabelsOutput)(nil), errors.New("labels failed"))
	_, err := service.DetectLabels(context.Background(), &visualanalysis.DetectLabelsInput{})
	assert.Error(t, err)
}

func TestVisualAnalysisCompareFaces(t *testing.T) {
	client := new(mockRekognitionClient)
	service := VisualAnalysis{RekognitionClient: client}
	client.On("CompareFaces", mock.Anything, mock.MatchedBy(func(in *rekognition.CompareFacesInput) bool {
		return aws.ToString(in.SourceImage.S3Object.Bucket) == "source-bucket" &&
			aws.ToString(in.TargetImage.S3Object.Name) == "target-key" &&
			aws.ToFloat32(in.SimilarityThreshold) == 85
	})).Return(&rekognition.CompareFacesOutput{
		FaceMatches: []rekognitiontypes.CompareFacesMatch{{
			Similarity: aws.Float32(92.5),
			Face: &rekognitiontypes.ComparedFace{
				Confidence:  aws.Float32(99),
				BoundingBox: &rekognitiontypes.BoundingBox{Left: aws.Float32(0.1), Top: aws.Float32(0.2)},
			},
		}},
		UnmatchedFaces: []rekognitiontypes.ComparedFace{{Confidence: aws.Float32(50)}},
	}, nil)

	result, err := service.CompareFaces(context.Background(), &visualanalysis.CompareFacesInput{
		SourceImage:         visualanalysis.ImageObjectRef{Bucket: "source-bucket", Key: "source-key"},
		TargetImage:         visualanalysis.ImageObjectRef{Bucket: "target-bucket", Key: "target-key"},
		SimilarityThreshold: 85,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.MatchedFaces))
	assert.Equal(t, 1, result.UnmatchedCount)
	assert.Equal(t, float32(92.5), result.HighestSimilarity)
	assert.Equal(t, float32(0.1), result.MatchedFaces[0].Face.BoundingBox.Left)
}

func TestVisualAnalysisCompareFacesError(t *testing.T) {
	client := new(mockRekognitionClient)
	service := VisualAnalysis{RekognitionClient: client}
	client.On("CompareFaces", mock.Anything, mock.Anything).Return((*rekognition.CompareFacesOutput)(nil), errors.New("compare failed"))
	_, err := service.CompareFaces(context.Background(), &visualanalysis.CompareFacesInput{})
	assert.Error(t, err)
}

func TestVisualAnalysisNilInputsUseEmptyImageRefs(t *testing.T) {
	client := new(mockRekognitionClient)
	service := VisualAnalysis{RekognitionClient: client}

	client.On("DetectFaces", mock.Anything, mock.MatchedBy(func(in *rekognition.DetectFacesInput) bool {
		return aws.ToString(in.Image.S3Object.Bucket) == "" && aws.ToString(in.Image.S3Object.Name) == ""
	})).Return(&rekognition.DetectFacesOutput{}, nil).Once()
	faces, err := service.DetectFaces(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, faces.FaceCount)

	client.On("DetectLabels", mock.Anything, mock.MatchedBy(func(in *rekognition.DetectLabelsInput) bool {
		return aws.ToString(in.Image.S3Object.Bucket) == "" && aws.ToString(in.Image.S3Object.Name) == ""
	})).Return(&rekognition.DetectLabelsOutput{}, nil).Once()
	labels, err := service.DetectLabels(context.Background(), nil)
	assert.NoError(t, err)
	assert.Empty(t, labels.Labels)

	client.On("CompareFaces", mock.Anything, mock.MatchedBy(func(in *rekognition.CompareFacesInput) bool {
		return aws.ToString(in.SourceImage.S3Object.Bucket) == "" &&
			aws.ToString(in.TargetImage.S3Object.Name) == ""
	})).Return(&rekognition.CompareFacesOutput{}, nil).Once()
	comparison, err := service.CompareFaces(context.Background(), nil)
	assert.NoError(t, err)
	assert.Empty(t, comparison.MatchedFaces)
}
