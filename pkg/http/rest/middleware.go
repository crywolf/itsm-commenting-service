package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/user"
	pb "github.com/KompiTech/itsm-user-service/api/userservice"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type userKeyType int

var userKey userKeyType

// AddUserInfo is a middleware that stores info about invoking user in request context
func (s Server) AddUserInfo(next httprouter.Handle, us UserService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		userData, err := us.UserBasicInfo(r)
		if err != nil {
			s.logger.Error("AddUserInfo middleware: UserBasicInfo service failed:", zap.Error(err))
			errMsg := errors.WithMessage(err, "could not retrieve correct user info from user service").Error()
			s.JSONError(w, errMsg, http.StatusInternalServerError)
			return
		}

		if userData.UUID == "" {
			s.logger.Error(fmt.Sprintf("AddUserInfo middleware: UserBasicInfo service returned invalid data: %v", userData))
			errMsg := errors.WithMessage(err, "could not retrieve correct user info from user service").Error()
			s.JSONError(w, errMsg, http.StatusInternalServerError)
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

// NewUserService creates user service with initialized GRPC client
func NewUserService() (UserService, error) {
	viper.SetDefault("UserServiceGRPCDialTarget", "localhost:50051")
	_ = viper.BindEnv("UserServiceGRPCDialTarget", "USER_SERVICE_GRPC_DIAL_TARGET")

	conn, err := grpc.Dial(viper.GetString("UserServiceGRPCDialTarget"),
		grpc.WithInsecure(),
	)

	if err != nil {
		return nil, err
	}

	return &userService{
		client: pb.NewUserManagementServiceClient(conn),
	}, nil
}

type userService struct {
	client pb.UserManagementServiceClient
}

// UserBasicInfo calls external use service and returns basic info about user who initiated the request
func (s userService) UserBasicInfo(r *http.Request) (user.BasicInfo, error) {
	md := metadata.New(map[string]string{
		"grpc-metadata-space": r.Header.Get("grpc-metadata-space"),
		"authorization":       r.Header.Get("authorization"),
	})

	ctx := metadata.NewOutgoingContext(context.Background(), md)

	var resp *pb.UserPersonalDetailsResponse
	resp, err := s.client.GetMyUserPersonalDetails(ctx, &emptypb.Empty{})
	if err != nil {
		return user.BasicInfo{}, err
	}

	u := resp.GetResult()

	userData := user.BasicInfo{
		UUID:           u.Uuid,
		Name:           u.Name,
		Surname:        u.Surname,
		OrgName:        u.OrgName,
		OrgDisplayName: u.OrgDisplayName,
	}

	return userData, nil
}
