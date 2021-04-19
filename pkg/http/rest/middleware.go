package rest

import (
	"context"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/user"
	"github.com/julienschmidt/httprouter"
)

type userKeyType int

var userKey userKeyType

// AddUserData is a middleware that stores info about invoking user in request context
func AddUserData(next httprouter.Handle, us user.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		userData, err := us.UserData(r)
		if err != nil {
			// TODO - err service unavailable? + log
		}

		ctx := context.WithValue(r.Context(), userKey, &userData)

		next(w, r.WithContext(ctx), ps)
	}
}

// UserFromContext returns the InvokingUserData value stored in ctx, if any.
func UserFromContext(ctx context.Context) (*user.InvokingUserData, bool) {
	u, ok := ctx.Value(userKey).(*user.InvokingUserData)
	return u, ok
}
