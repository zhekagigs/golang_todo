package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/zhekagigs/golang_todo/logger"
)

type LoggingMiddleware struct {
	Next http.Handler
}

func (m LoggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	crw := &customResponseWriter{ResponseWriter: w} // TODO remove

	m.Next.ServeHTTP(crw, r)

	duration := time.Since(start)

	logger.Info.Printf(
		"Method: %s, Path: %s, Status: %d, Duration: %v",
		r.Method,
		r.URL.Path,
		crw.status,
		duration,
	)
}

type userKey struct{}

// a real implementation would be signed to make sure
// the user didn't spoof their identity
func extractUser(req *http.Request) (string, error) {
	if strings.Contains(req.URL.Path, "api") {
		identityId, ok := req.Header["Authorization"]
		if !ok || len(identityId) == 0 || identityId[0] == "" {
			return "", errors.New("no identity header found")
		}
		return identityId[0], nil
	} else {
		return extractUserFromCookie(req)
	}

}

func extractUserFromCookie(req *http.Request) (string, error) {
	identityId, err := req.Cookie("Authorization")
	if err != nil || len(identityId.Value) == 0 {
		return "", err
	}
	return identityId.Value, nil
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userVal, err := extractUser(r)

		if err != nil || userVal == "" {
			logger.Error.Println("error extacting user id", err)
			w.Write([]byte("<div>Please go back and login</div> <a href=/tasks> Main Page</a>"))
			// http.Redirect(w, r, "/tasks", http.StatusUnauthorized)
			return
		}
		ctx := r.Context()
		ctx = ContextWithUser(ctx, userVal)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}
}

func ContextWithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

func UserFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(userKey{}).(string)
	return user, ok
}

type customResponseWriter struct {
	http.ResponseWriter
	status int
}
