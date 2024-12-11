package internal

import (
	"context"
	"database/sql"
	"testing"

	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

var db *sql.DB

// TODO move testcases to another file ?
var registerCases = []struct {
	name      string
	request   *pb.RegisterRequest
	expectErr error
}{
	{
		name: "success",
		request: &pb.RegisterRequest{
			Username:    "foouser",
			Email:       "foofoo@mail.com",
			Password:    "secretPass12.",
			PhoneNumber: "0812345678",
		},
		expectErr: nil,
	},
}

func TestRegister(t *testing.T) {

	sv := NewAuthService(mockStorage{}, mockProfileClient{})

	for _, c := range registerCases {

		if c.name != "success" {
			t.Run("fail case", func(t *testing.T) {
			})
			continue
		}

		t.Run(c.name, func(t *testing.T) {

			users, err := sv.Register(context.TODO(), c.request)
			if err != nil {
				t.Errorf("sv.Register(ctx,c.request) = %v, want nil", err)
			}

			_ = users

		})

	}
}

func TestLogin(t *testing.T) {
}
