package server

import (
	"command-on-demand/internal/errors"
	"command-on-demand/internal/logger"
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/dchest/uniuri"
)

type ctxKey string

const ctxKeyRequestId ctxKey = "reqId"

func getRequestId(r *http.Request) string {
	ctx := r.Context()
	val := ctx.Value(ctxKeyRequestId)

	if rId, ok := val.(string); ok {
		return rId
	}

	return ""
}

func MiddlewareSetRequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rId := uniuri.NewLen(10)
		r = r.WithContext(context.WithValue(r.Context(), ctxKeyRequestId, rId))
		next.ServeHTTP(w, r)
	})
}

func MiddlewareLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rId := getRequestId(r)

		logger.WithRequest(rId, r).Info("request received")

		next.ServeHTTP(w, r)
	})
}

func (s Server) MiddlewareBearerAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rId := getRequestId(r)

		t := r.Header.Get("Authorization")

		if !strings.HasPrefix(t, "Bearer ") {
			logger.WithRequest(rId, r).Error(errors.BadToken)
			writeErrorResponse(w, errors.BadToken)
			return
		}

		// convert to []byte outside of CTC as these are not constant time ops
		bt := []byte(strings.TrimPrefix(t, "Bearer "))
		bst := []byte(s.token())

		match := subtle.ConstantTimeCompare(bt, bst)
		if match != 1 {
			logger.WithRequest(rId, r).Error(errors.InvalidToken)
			writeErrorResponse(w, errors.InvalidToken)
			return
		}

		logger.WithRequest(rId, r).Info("token authentication successful")
		next.ServeHTTP(w, r)
	})
}
