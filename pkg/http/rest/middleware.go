package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/user"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

type userKeyType int

var userKey userKeyType

// AddUserData is a middleware that stores info about invoking user in request context
func (s Server) AddUserData(next httprouter.Handle, us user.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		userData, err := us.UserData(r)
		if err != nil {
			s.logger.Error("AddUserData middleware: UserData service failed:", zap.Error(err))
			s.JSONError(w, "could not retrieve correct user info from user service", http.StatusInternalServerError)
			return
		}

		if userData.UUID == "" {
			s.logger.Error(fmt.Sprintf("AddUserData middleware: UserData service returned invalid data: %v", userData))
			s.JSONError(w, "could not retrieve correct user info from user service", http.StatusInternalServerError)
			return
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
