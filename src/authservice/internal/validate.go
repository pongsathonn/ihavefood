package internal

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errUserIncorrect   = status.Error(codes.InvalidArgument, "username or password incorrect")
	errUserNotFound    = status.Error(codes.NotFound, "user not found")
	errPasswordHashing = status.Error(codes.Internal, "password hashing failed")

	errNoToken       = status.Error(codes.InvalidArgument, "token must be provided")
	errInvalidToken  = status.Error(codes.Unauthenticated, "invalid token")
	errGenerateToken = status.Error(codes.Internal, "failed to generate authentication token")
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func validateUser(in any) error {

	registerRule := map[string]string{
		"Username":    "required,min=6,max=16,lowercase",
		"Email":       "required,email",
		"Password":    "required,vfpass,min=8,max=16",
		"PhoneNumber": "required,vfphone",
	}

	validate.RegisterStructValidationMapRules(registerRule, pb.RegisterRequest{})
	// validate.RegisterStructValidationMapRules(rule2, nil)

	// prefix vf = validate format
	// Ex. vfpass validates password format
	validate.RegisterValidation("vfpass", validatePassword)
	validate.RegisterValidation("vfphone", validatePhone)

	if err := validate.Struct(in); err != nil {
		var errs []string
		for _, v := range err.(validator.ValidationErrors) {
			var e error
			switch v.Tag() {
			case "required":
				e = fmt.Errorf("%s must be provided", v.Field())
			case "email":
				e = errors.New("invalid email")
			case "vfpass":
				e = errors.New("password must contain lowercase,uppercase and special character")
			case "vfphone":
				e = errors.New("invalid phone number format")
			case "min":
				e = fmt.Errorf("%s must be at least %s", v.Field(), v.Param())
			case "max":
				e = fmt.Errorf("%s must be at most %s", v.Field(), v.Param())
			case "lowercase":
				e = fmt.Errorf("%s must be lowercase only", v.Field())
			}
			errs = append(errs, e.Error())
		}
		return errors.New(strings.Join(errs, ", "))
	}
	return nil
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return regexp.MustCompile(`[a-z]`).MatchString(password) &&
		regexp.MustCompile(`[A-Z]`).MatchString(password) &&
		regexp.MustCompile(`[!.@#$%^&*()_\-+=<>?]`).MatchString(password)
}

func validatePhone(fl validator.FieldLevel) bool {
	phoneNumber := fl.Field().String()
	// phone number format (e.g., 06XXXXXXXX, 08XXXXXXXX, 09XXXXXXXX).
	// Any format outside of this is considered invalid, and the function
	// returns an error.
	return regexp.MustCompile(`^(06|08|09)\d{8}$`).MatchString(phoneNumber)
}

// verifyUserToken verifies the validity of a JWT token using the provided signing key.
// It returns true if the token is valid, false otherwise, along with any error encountered.
func verifyUserToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})
	if err != nil {
		return false, err
	}

	if !token.Valid {
		return false, errors.New("invalid user token")
	}

	return true, nil
}

func verifyAdminToken(tokenString string) (bool, error) {
	token, err := jwt.ParseWithClaims(tokenString, new(AuthClaims), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return false, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})
	if err != nil {
		return false, err
	}

	if claims, _ := token.Claims.(*AuthClaims); claims.Role != pb.Roles_ADMIN {
		return false, errors.New("token claims do not have admin role")
	}

	return true, nil
}
