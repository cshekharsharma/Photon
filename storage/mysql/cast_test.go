package mysql

import (
	"database/sql"
	"testing"
	"time"
)

func TestCastSQLStringToString(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{"Null case", sql.NullString{String: "", Valid: false}, ""},
		{"Valid case", sql.NullString{String: "hello", Valid: true}, "hello"},
		{"Valid cache case", map[string]interface{}{"String": "hello", "Valid": true}, "hello"},
		{"Invalid cache case", map[string]interface{}{"String": 123, "Valid": true}, ""},
		{"Unknown type", 123, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CastSQLStringToString(tt.value); got != tt.want {
				t.Errorf("CastSQLStringToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCastSQLStringToInt(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  int64
	}{
		{"Null case", sql.NullString{String: "", Valid: false}, 0},
		{"Valid integer", sql.NullString{String: "123", Valid: true}, 123},
		{"Valid cache case", map[string]interface{}{"String": "123", "Valid": true}, 123},
		{"invalid cache case", map[string]interface{}{"String": "abc", "Valid": true}, 0},
		{"cache valid missing string", map[string]interface{}{"Valid": true}, 0},
		{"cache valid wrong flag type", map[string]interface{}{"String": "123", "Valid": "true"}, 0},
		{"Invalid integer", sql.NullString{String: "abc", Valid: true}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CastSQLStringToInt(tt.value); got != tt.want {
				t.Errorf("CastSQLStringToInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCastSQLStringToFloat(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  float64
	}{
		{"Null case", sql.NullString{String: "", Valid: false}, 0.0},
		{"Valid float", sql.NullString{String: "123.456", Valid: true}, 123.456},
		{"Valid float case", map[string]interface{}{"String": "123.456", "Valid": true}, 123.456},
		{"invalid cache case", map[string]interface{}{"String": "abc", "Valid": true}, 0.0},
		{"cache valid missing string", map[string]interface{}{"Valid": true}, 0.0},
		{"cache valid wrong flag type", map[string]interface{}{"String": "123.4", "Valid": "true"}, 0.0},
		{"Invalid float", sql.NullString{String: "abc", Valid: true}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CastSQLStringToFloat(tt.value); got != tt.want {
				t.Errorf("CastSQLStringToFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCastSQLStringToTime(t *testing.T) {
	defaultTime, _ := time.Parse(time.DateTime, "")
	validTime, _ := time.Parse(time.DateTime, "2023-01-01 00:00:00")

	tests := []struct {
		name  string
		value interface{}
		want  time.Time
	}{
		{"Null case", sql.NullString{String: "", Valid: false}, defaultTime},
		{"Valid time", sql.NullString{String: "2023-01-01 00:00:00", Valid: true}, validTime},
		{"Valid cache case", map[string]interface{}{"String": "2023-01-01 00:00:00", "Valid": true}, validTime},
		{"invalid cache time case", map[string]interface{}{"String": "not-a-time", "Valid": true}, defaultTime},
		{"cache missing string", map[string]interface{}{"Valid": true}, defaultTime},
		{"cache invalid flag type", map[string]interface{}{"String": "2023-01-01 00:00:00", "Valid": "true"}, defaultTime},
		{"Invalid time", sql.NullString{String: "not-a-time", Valid: true}, defaultTime},
		{"Cache invalid flag", map[string]interface{}{"String": "2023-01-01 00:00:00", "Valid": false}, defaultTime},
		{"Unknown type", 123, defaultTime},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CastSQLStringToTime(tt.value); !got.Equal(tt.want) {
				t.Errorf("CastSQLStringToTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
