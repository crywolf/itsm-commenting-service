package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/user"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/usersvc"
	grpc2http "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

type userKeyType int

var userKey userKeyType

// AddUserInfo is a middleware that stores info about invoking user in request context
// (or about user this request is made on behalf of)
func (s Server) AddUserInfo(next httprouter.Handle, us usersvc.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		userData, err := us.UserBasicInfo(r)
		if err != nil {
			s.logger.Error("AddUserInfo middleware: UserBasicInfo service failed:", zap.Error(err))
			errMsg := errors.WithMessage(err, "could not retrieve correct user info from user service").Error()
			statusCode := grpc2http.HTTPStatusFromCode(status.Code(err))
			s.JSONError(w, errMsg, statusCode)
			return
		}

		if userData.UUID == "" {
			s.logger.Error(fmt.Sprintf("AddUserInfo middleware: UserBasicInfo service returned invalid data: %v", userData))
			errMsg := errors.WithMessage(err, "could not retrieve correct user info from user service").Error()
			statusCode := grpc2http.HTTPStatusFromCode(status.Code(err))
			s.JSONError(w, errMsg, statusCode)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, &userData)

		next(w, r.WithContext(ctx), ps)
	}
}

// UserInfoFromContext returns the BasicInfo value stored in ctx, if any.
func (s Server) UserInfoFromContext(ctx context.Context) (*user.BasicInfo, bool) {
	u, ok := ctx.Value(userKey).(*user.BasicInfo)
	return u, ok
}
