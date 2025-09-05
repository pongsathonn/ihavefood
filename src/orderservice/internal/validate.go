package internal

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
)

func SetupValidator() {

	validate.RegisterStructValidationMapRules(map[string]string{
		"RequestId":         "required,uuid4",
		"CustomerId":        "required,uuid4",
		"MerchantId":        "required,uuid4",
		"Items":             "required,vitems",
		"CustomerAddressId": "required,uuid4",
		"PaymentMethods":    "vpayment_method",
	}, pb.CreatePlaceOrderRequest{})

	err := validate.RegisterValidation("vitems", func(fl validator.FieldLevel) bool {
		items, ok := fl.Field().Interface().([]*pb.OrderItem)
		if !ok || len(items) == 0 {
			return false
		}

		for _, item := range items {

			// Items.ItemId must be valid uuid4
			u, err := uuid.Parse(item.ItemId)
			if err != nil || u.Version() != 4 {
				return false
			}

			// Items.Quantity must be at least 1
			if item.Quantity < 1 {
				return false
			}
		}

		return true
	})

	err = validate.RegisterValidation("vphone", func(fl validator.FieldLevel) bool {
		phoneNumber := fl.Field().String()
		return regexp.MustCompile(`^(06|08|09)\d{8}$`).MatchString(phoneNumber)
	})
	if err != nil {
		log.Fatalf("unable to register vphone: %v", err)
	}

	err = validate.RegisterValidation("vpayment_method", func(fl validator.FieldLevel) bool {
		value := fl.Field().Interface().(pb.PaymentMethods)
		if value == pb.PaymentMethods_PAYMENT_METHOD_UNSPECIFIED {
			return false
		}
		_, exists := pb.PaymentMethods_value[value.String()]
		return exists
	})
	if err != nil {
		log.Fatalf("unable to register vpayment_method: %v", err)
	}

	slog.Info("Order service validator initialized")
}

var validate = validator.New(validator.WithRequiredStructEnabled())

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
	case "uuid4":
		return myValidatorErr{Field: f.Field(), Msg: "must be a valid UUID4"}
	case "min":
		return myValidatorErr{Field: f.Field(), Msg: fmt.Sprintf("must be at least %s", f.Param())}
	case "vpayment_method":
		var methods []string
		for method := range pb.PaymentMethods_value {
			if method != pb.PaymentMethods_PAYMENT_METHOD_UNSPECIFIED.String() {
				methods = append(methods, method)
			}
		}
		return myValidatorErr{
			Field: f.Field(),
			Msg:   fmt.Sprintf("must be one of %s", strings.Join(methods, ", ")),
		}
	case "vitems":
		return myValidatorErr{Field: f.Field(), Msg: "invalid: itemId must be a valid UUID4 and quantity must be at least 1"}
	default:
		return myValidatorErr{Field: f.Field(), Msg: fmt.Sprintf("invalid value tag %s", f.Tag())}
	}
}
