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

// AddUserInfo is a middleware that stores info about invoking user in request context
func (s Server) AddUserInfo(next httprouter.Handle, us UserService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		userData, err := us.UserBasicInfo(r)
		if err != nil {
			s.logger.Error("AddUserInfo middleware: UserBasicInfo service failed:", zap.Error(err))
			s.JSONError(w, "could not retrieve correct user info from user service", http.StatusInternalServerError)
			return
		}

		if userData.UUID == "" {
			s.logger.Error(fmt.Sprintf("AddUserInfo middleware: UserBasicInfo service returned invalid data: %v", userData))
			s.JSONError(w, "could not retrieve correct user info from user service", http.StatusInternalServerError)
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

// UserService provides basic info about user
type UserService interface {
	UserBasicInfo(r *http.Request) (user.BasicInfo, error)
}

// NewUserService creates user service
func NewUserService() UserService {
	return &userService{}
}

// userService calls external user service that provides basic info about user
type userService struct {
	//client
}

// UserBasicInfo returns basic info about user who initiated the request
func (s userService) UserBasicInfo(r *http.Request) (user.BasicInfo, error) {
	//authHeader := r.Header.Get("authorization")
	//fmt.Println(authHeader)

	// TODO fetch real user from some user service
	userData := user.BasicInfo{
		UUID:           "2af4f493-0bd5-4513-b440-6cbb465feadb",
		Name:           "Alice",
		Surname:        "Cooper",
		OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
		OrgDisplayName: "Kompitech",
	}

	return userData, nil
}
