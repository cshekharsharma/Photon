package qdrant

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestPointIDMarshalJSON(t *testing.T) {
	id := PointIDString("abc")
	b, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("marshal string id: %v", err)
	}
	if string(b) != "\"abc\"" {
		t.Fatalf("unexpected string id json: %s", string(b))
	}

	id = PointIDInt(123)
	b, err = json.Marshal(id)
	if err != nil {
		t.Fatalf("marshal int id: %v", err)
	}
	if string(b) != "123" {
		t.Fatalf("unexpected int id json: %s", string(b))
	}

	var invalid PointID
	if _, err := json.Marshal(invalid); err == nil {
		t.Fatalf("expected error for invalid point id")
	}
}

func TestPointID_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string // Raw JSON input
		want    PointID
		wantErr bool
	}{
		// --- Valid String IDs ---
		{
			name: "valid string UUID",
			json: `"50e93080-04bd-44a3-863a-230006778f01"`,
			want: PointID{kind: PointIDTypeString, s: "50e93080-04bd-44a3-863a-230006778f01"},
		},
		{
			name: "simple string id",
			json: `"my_point"`,
			want: PointID{kind: PointIDTypeString, s: "my_point"},
		},
		{
			name: "numeric string (quotes make it string)",
			json: `"12345"`,
			want: PointID{kind: PointIDTypeString, s: "12345"},
		},
		{
			name: "string with escaped quotes",
			json: `"id_with_\"quotes\""`,
			want: PointID{kind: PointIDTypeString, s: `id_with_"quotes"`},
		},

		// --- Valid Integer IDs ---
		{
			name: "positive integer",
			json: `12345`,
			want: PointID{kind: PointIDTypeInt, i: 12345},
		},
		{
			name: "zero integer",
			json: `0`,
			want: PointID{kind: PointIDTypeInt, i: 0},
		},
		{
			name: "negative integer",
			json: `-999`,
			want: PointID{kind: PointIDTypeInt, i: -999},
		},
		{
			name: "large integer (int64 max)",
			json: `9223372036854775807`,
			want: PointID{kind: PointIDTypeInt, i: 9223372036854775807},
		},

		// --- Null / Empty ---
		{
			name: "null input",
			json: `null`,
			want: PointID{}, // Zero value
		},
		{
			name: "empty byte slice",
			json: ``,
			want: PointID{},
		},

		// --- Invalid Inputs ---
		{
			name:    "float (invalid for ID)",
			json:    `123.45`,
			wantErr: true, // strconv.ParseInt fails on '.'
		},
		{
			name:    "boolean true",
			json:    `true`,
			wantErr: true,
		},
		{
			name:    "boolean false",
			json:    `false`,
			wantErr: true,
		},
		{
			name:    "json object",
			json:    `{"id": 1}`,
			wantErr: true,
		},
		{
			name:    "json array",
			json:    `[1, 2]`,
			wantErr: true,
		},
		{
			name:    "garbage data",
			json:    `invalid_json_text`,
			wantErr: true,
		},
		{
			name:    "malformed string",
			json:    `"unclosed string`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id PointID
			err := id.UnmarshalJSON([]byte(tt.json))

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error expected, check the value
			if !tt.wantErr && !reflect.DeepEqual(id, tt.want) {
				t.Errorf("UnmarshalJSON() got = %+v, want %+v", id, tt.want)
			}
		})
	}
}
