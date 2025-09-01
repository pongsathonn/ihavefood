package internal

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

type myValidatorErrs []myValidatorErr

func (m myValidatorErrs) Error() string {
	var s []string
	for _, err := range m {
		s = append(s, err.Error())
	}
	return strings.Join(s, ", ")
}

type myValidatorErr struct {
	Field string
	Msg   string
}

func (m myValidatorErr) Error() string {
	return fmt.Sprintf("%s %s", m.Field, m.Msg)
}

var validate = validator.New(validator.WithRequiredStructEnabled())

func SetupValidator() {
	validate.RegisterStructValidationMapRules(map[string]string{
		"Username":    "required,min=6,max=16,lowercase",
		"Email":       "required,email",
		"Password":    "required,vfpass,min=8,max=16",
		"PhoneNumber": "required,vfphone",
	}, pb.RegisterRequest{})
	// validate.RegisterStructValidationMapRules(rule2, nil)

	// prefix vf = validate format
	// Ex. vfpass validates password format
	validate.RegisterValidation("vfpass", validatePassword)
	validate.RegisterValidation("vfphone", validatePhone)
}

func ValidateStruct(in any) error {
	if err := validate.Struct(in); err != nil {

		var valErrs validator.ValidationErrors
		if !errors.As(err, &valErrs) {
			return err
		}

		var errs myValidatorErrs
		for _, valErr := range valErrs {
			errs = append(errs, buildMyValidatorErr(valErr))
		}
		return errs
	}
	return nil
}

func buildMyValidatorErr(f validator.FieldError) myValidatorErr {
	switch f.Tag() {
	case "required":
		return myValidatorErr{Field: f.Field(), Msg: "is required"}
	case "email":
		return myValidatorErr{Field: f.Field(), Msg: "must be a valid email address"}
	case "vfpass":
		return myValidatorErr{Field: f.Field(), Msg: "must contain lowercase,uppercase and special character"}
	case "vfphone":
		return myValidatorErr{Field: f.Field(), Msg: "must be a valid phone number format"}
	case "min":
		return myValidatorErr{Field: f.Field(), Msg: fmt.Sprintf("must be at least %s", f.Param())}
	case "max":
		return myValidatorErr{Field: f.Field(), Msg: fmt.Sprintf("must be at most %s", f.Param())}
	case "lowercase":
		return myValidatorErr{Field: f.Field(), Msg: "must be lowercase only"}
	default:
		return myValidatorErr{Field: f.Field(), Msg: fmt.Sprintf("invalid value tag %s", f.Tag())}
	}
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
