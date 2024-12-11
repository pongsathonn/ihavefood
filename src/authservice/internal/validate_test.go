package internal

import (
	"testing"

	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

var validateCases = []struct {
	name    string
	request *pb.RegisterRequest
	want    string
}{
	{
		name: "valid",
		request: &pb.RegisterRequest{
			Username:    "farboo",
			Email:       "farboo@example.com",
			Password:    "$ecretX1.",
			PhoneNumber: "0987654321",
		},
		want: "register successful",
	},
	{
		name: "valid",
		request: &pb.RegisterRequest{
			Username:    "kukugaga2",
			Email:       "farboo@example.com",
			Password:    "passWould2@.",
			PhoneNumber: "0887654321",
		},
		want: "register successful",
	},
	{
		name: "empty request fields",
		request: &pb.RegisterRequest{
			Username:    "",
			Email:       "",
			Password:    "",
			PhoneNumber: "",
		},
		want: "err: fields must be provided",
	},
	{
		name: "invalid username format",
		request: &pb.RegisterRequest{
			Username:    "tEstUsER9",
			Email:       "testuser1@mail.com",
			Password:    "skkknsW21.",
			PhoneNumber: "0671654293",
		},
		want: "err: username must be lowercase",
	},
	{
		name: "invalid minimum length",
		request: &pb.RegisterRequest{
			Username:    "user",
			Email:       "testuser1@mail.com",
			Password:    "$eC2et.",
			PhoneNumber: "0671654293",
		},
		want: "err: invalid username and password minimum length ",
	},
	{
		name: "invalid maximum length",
		request: &pb.RegisterRequest{
			Username:    "usertestusertestusertest",
			Email:       "testuser1@mail.com",
			Password:    "$eC2et.asijd92@nlax80.as2as9o",
			PhoneNumber: "0671654293",
		},
		want: "err: invalid username and password maximum length ",
	},
	{
		name: "invalid format email, password, phone number",
		request: &pb.RegisterRequest{
			Username:    "testuser",
			Email:       "invalid-email",
			Password:    "short",
			PhoneNumber: "++66987691765",
		},
		want: "err: email, password or phone number invalid format",
	},
}

func TestValidateUser(t *testing.T) {

	for _, c := range validateCases {

		if c.name != "valid" {
			t.Run(c.name, func(t *testing.T) {
				if err := validateUser(c.request); err == nil {
					t.Errorf("error = nil, want %s", c.want)
				}
			})
			continue
		}

		// valid cases
		t.Run(c.name, func(t *testing.T) {
			if err := validateUser(c.request); err != nil {
				t.Errorf("error = %v, want nil", err)
			}
		})

	}

}
