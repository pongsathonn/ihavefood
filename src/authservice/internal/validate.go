package internal

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
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
		"Email":       "required,email",
		"Password":    "required,vpass,min=8,max=16",
		"PhoneNumber": "required,vphone",
		"Role":        "vrole",
	}, pb.RegisterRequest{})

	validate.RegisterStructValidationMapRules(map[string]string{
		"Identifier": "required",
		"Password":   "required",
	}, pb.LoginRequest{})

	// validate.RegisterStructValidationMapRules(rule2, nil)

	// prefix 'v' for custom validation.
	// Examples:
	// - vpass validates password format.
	// - vrole validates roles enum.
	validate.RegisterValidation("vpass", validatePassword)
	validate.RegisterValidation("vphone", validatePhone)
	validate.RegisterValidation("vrole", validateRole)
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
	case "min":
		return myValidatorErr{Field: f.Field(), Msg: fmt.Sprintf("must be at least %s", f.Param())}
	case "max":
		return myValidatorErr{Field: f.Field(), Msg: fmt.Sprintf("must be at most %s", f.Param())}
	case "lowercase":
		return myValidatorErr{Field: f.Field(), Msg: "must be lowercase only"}
	case "vpass":
		return myValidatorErr{Field: f.Field(), Msg: "must contain lowercase,uppercase and special character"}
	case "vphone":
		return myValidatorErr{Field: f.Field(), Msg: "must be a valid phone number format"}
	case "vrole":
		var roles []string
		for role := range pb.Roles_value {
			if role != pb.Roles_ROLES_UNSPECIFIED.String() {
				roles = append(roles, role)
			}
		}
		return myValidatorErr{
			Field: f.Field(),
			Msg:   fmt.Sprintf("must be one of %s", strings.Join(roles, ", ")),
		}
	default:
		return myValidatorErr{Field: f.Field(), Msg: fmt.Sprintf("invalid value tag %s", f.Tag())}
	}
}

func validateRole(fl validator.FieldLevel) bool {

	r := fl.Field().Interface().(pb.Roles)
	if r == pb.Roles_ROLES_UNSPECIFIED {
		return false
	}

	_, exists := pb.Roles_value[r.String()]
	return exists
}

// phone number format (e.g., 06XXXXXXXX, 08XXXXXXXX, 09XXXXXXXX).
// Any format outside of this is considered invalid, and the function
// returns an error.
func validatePhone(fl validator.FieldLevel) bool {
	phoneNumber := fl.Field().String()
	return regexp.MustCompile(`^(06|08|09)\d{8}$`).MatchString(phoneNumber)
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return regexp.MustCompile(`[a-z]`).MatchString(password) &&
		regexp.MustCompile(`[A-Z]`).MatchString(password) &&
		regexp.MustCompile(`[!.@#$%^&*()_\-+=<>?]`).MatchString(password)
}
