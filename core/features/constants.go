package features

type FeatureSourceType uint8 // FeatureSourceType represents the source type of the feature config.

const (
	WildCardValue = "*"

	ConjunctionAnd = "AND"
	ConjunctionOr  = "OR"

	OperatorEquals             = "EQUALS"
	OperatorNotEquals          = "NOT_EQUALS"
	OperatorGreaterThan        = "GREATER_THAN"
	OperatorLessThan           = "LESS_THAN"
	OperatorGreaterThanOrEqual = "GREATER_THAN_OR_EQUAL"
	OperatorLessThanOrEqual    = "LESS_THAN_OR_EQUAL"
	OperatorIn                 = "IN"
	OperatorNotIn              = "NOT_IN"

	DatatypeString  = "STRING"
	DatatypeInteger = "INTEGER"
	DatatypeFloat   = "FLOAT"
	DatatypeBoolean = "BOOLEAN"

	FeatureSourceFile     FeatureSourceType = iota // when feature config is loaded from a file.
	FeatureSourceRawBytes                          // when feature config is provided as raw JSON content.

	FeatureContentFormatJson = 1
)

var validConjunctions = []string{
	ConjunctionAnd,
	ConjunctionOr,
}

var validOperators = []string{
	OperatorEquals,
	OperatorNotEquals,
	OperatorGreaterThan,
	OperatorLessThan,
	OperatorGreaterThanOrEqual,
	OperatorLessThanOrEqual,
	OperatorIn,
	OperatorNotIn,
}

var validDatatypes = []string{
	DatatypeString,
	DatatypeInteger,
	DatatypeFloat,
	DatatypeBoolean,
}
