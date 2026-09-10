package apiresponse

import (
	"net/http"
)

type InternalResponseCode string

type InternalResponse struct {
	httpCode int
	message  string
}

var codesMap = map[InternalResponseCode]InternalResponse{
	AllOk: {
		httpCode: http.StatusOK,
		message:  http.StatusText(http.StatusOK),
	},
	Created: {
		httpCode: http.StatusCreated,
		message:  http.StatusText(http.StatusCreated),
	},
	Accepted: {
		httpCode: http.StatusAccepted,
		message:  http.StatusText(http.StatusAccepted),
	},
	NoContent: {
		httpCode: http.StatusNoContent,
		message:  http.StatusText(http.StatusNoContent),
	},
	ResetContent: {
		httpCode: http.StatusResetContent,
		message:  http.StatusText(http.StatusResetContent),
	},
	PartialContent: {
		httpCode: http.StatusPartialContent,
		message:  http.StatusText(http.StatusPartialContent),
	},
	MovedPermanently: {
		httpCode: http.StatusMovedPermanently,
		message:  http.StatusText(http.StatusMovedPermanently),
	},
	Found: {
		httpCode: http.StatusFound,
		message:  http.StatusText(http.StatusFound),
	},
	RedirectTemporarily: {
		httpCode: http.StatusTemporaryRedirect,
		message:  http.StatusText(http.StatusTemporaryRedirect),
	},
	RedirectPermanently: {
		httpCode: http.StatusPermanentRedirect,
		message:  http.StatusText(http.StatusPermanentRedirect),
	},
	BadRequest: {
		httpCode: http.StatusBadRequest,
		message:  http.StatusText(http.StatusBadRequest),
	},
	Forbidden: {
		httpCode: http.StatusForbidden,
		message:  http.StatusText(http.StatusForbidden),
	},
	NotFound: {
		httpCode: http.StatusNotFound,
		message:  http.StatusText(http.StatusNotFound),
	},
	NotAcceptable: {
		httpCode: http.StatusNotAcceptable,
		message:  http.StatusText(http.StatusNotAcceptable),
	},
	RequestTimeout: {
		httpCode: http.StatusRequestTimeout,
		message:  http.StatusText(http.StatusRequestTimeout),
	},
	Conflict: {
		httpCode: http.StatusConflict,
		message:  http.StatusText(http.StatusConflict),
	},
	UnsupportedMediaType: {
		httpCode: http.StatusUnsupportedMediaType,
		message:  http.StatusText(http.StatusUnsupportedMediaType),
	},
	UnprocessableEntity: {
		httpCode: http.StatusUnprocessableEntity,
		message:  http.StatusText(http.StatusUnprocessableEntity),
	},
	TooManyRequests: {
		httpCode: http.StatusTooManyRequests,
		message:  http.StatusText(http.StatusTooManyRequests),
	},
	InternalServerError: {
		httpCode: http.StatusInternalServerError,
		message:  http.StatusText(http.StatusInternalServerError),
	},
	BadGateway: {
		httpCode: http.StatusBadGateway,
		message:  http.StatusText(http.StatusBadGateway),
	},
	ServiceUnavailable: {
		httpCode: http.StatusServiceUnavailable,
		message:  http.StatusText(http.StatusServiceUnavailable),
	},
	GatewayTimeout: {
		httpCode: http.StatusGatewayTimeout,
		message:  http.StatusText(http.StatusGatewayTimeout),
	},
	Unauthorized: {
		httpCode: http.StatusUnauthorized,
		message:  http.StatusText(http.StatusUnauthorized),
	},
}

// Application specific API response codes.
// General convention is to have "S" prefix to success codes,
// and "E" perfix to error/failure codes.
// Important to note that, these code are and should be treated
// as completely independent of http response codes. EX: E500 does
// not imply that it will only be sent in case of 500 Server Error.

// Server panic.
const (
	AllOk                InternalResponseCode = "S200"
	Created              InternalResponseCode = "S201"
	Accepted             InternalResponseCode = "S202"
	NoContent            InternalResponseCode = "E204"
	ResetContent         InternalResponseCode = "E205"
	PartialContent       InternalResponseCode = "E206"
	MovedPermanently     InternalResponseCode = "E301"
	Found                InternalResponseCode = "E302"
	RedirectTemporarily  InternalResponseCode = "E307"
	RedirectPermanently  InternalResponseCode = "E308"
	BadRequest           InternalResponseCode = "E400"
	Unauthorized         InternalResponseCode = "E401"
	Forbidden            InternalResponseCode = "E403"
	NotFound             InternalResponseCode = "E404"
	NotAcceptable        InternalResponseCode = "E406"
	RequestTimeout       InternalResponseCode = "E408"
	Conflict             InternalResponseCode = "E409"
	UnsupportedMediaType InternalResponseCode = "E415"
	UnprocessableEntity  InternalResponseCode = "E422"
	TooManyRequests      InternalResponseCode = "E429"
	InternalServerError  InternalResponseCode = "E500"
	BadGateway           InternalResponseCode = "E502"
	ServiceUnavailable   InternalResponseCode = "E503"
	GatewayTimeout       InternalResponseCode = "E504"
)

// Translate application response code a valid HTTP response code
// that can be sent over the network to http clients.
func ToHttpCode(appcode InternalResponseCode) int {
	if item, ok := codesMap[appcode]; ok {
		return item.httpCode
	}

	return http.StatusOK
}
