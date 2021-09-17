package usersvc

import (
	"context"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/user"
	"github.com/KompiTech/itsm-user-service/api/userservice"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Service provides basic info about user
type Service interface {
	UserBasicInfo(r *http.Request) (user.BasicInfo, error)
}

// NewService creates user service with initialized GRPC client
func NewService() (Service, error) {
	conn, err := grpc.Dial(
		viper.GetString("UserServiceGRPCDialTarget"),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	return &userService{
		client: usermanagement.NewUserManagementServiceClient(conn),
	}, nil
}

type userService struct {
	client usermanagement.UserManagementServiceClient
}

// UserBasicInfo calls external use service and returns basic info about user who initiated the request
// or about user this request is made on behalf of
func (s userService) UserBasicInfo(r *http.Request) (user.BasicInfo, error) {
	md := metadata.New(map[string]string{
		"grpc-metadata-space": r.Header.Get("grpc-metadata-space"),
		"authorization":       r.Header.Get("authorization"),
	})

	ctx := metadata.NewOutgoingContext(context.Background(), md)

	var resp *usermanagement.UserPersonalDetailsResponse
	var err error

	if onBehalf := r.Header.Get("on_behalf"); onBehalf != "" {
		resp, err = s.client.UserGet(ctx, &usermanagement.UserRequest{Uuid: onBehalf})
		if err != nil {
			return user.BasicInfo{}, err
		}
	} else {
		resp, err = s.client.UserGetMyPersonalDetails(ctx, &emptypb.Empty{})
		if err != nil {
			return user.BasicInfo{}, err
		}
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
