package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

var ErrCustomerNotFound = errors.New("customer not found")

func NewCustomerStorage(db *sql.DB) *customerStorage {
	return &customerStorage{db: db}
}

type customerStorage struct {
	db *sql.DB
}

// Customers returns a list of customers.
func (s *customerStorage) customers(ctx context.Context) ([]*dbCustomer, error) {
	customerRows, err := s.db.QueryContext(ctx, `
		SELECT
			customer_id, username, bio, facebook, instagram, line, create_time, update_time
		FROM customers
		ORDER BY CAST(customer_id AS INTEGER)
	`)
	if err != nil {
		return nil, err
	}
	defer customerRows.Close()

	customersMap := make(map[string]*dbCustomer)
	var customerIDs []string

	for customerRows.Next() {
		var p dbCustomer
		err := customerRows.Scan(
			&p.UserID, &p.Username, &p.Bio, &p.Social.Facebook,
			&p.Social.Instagram, &p.Social.Line, &p.CreateTime, &p.UpdateTime,
		)
		if err != nil {
			return nil, err
		}
		p.Addresses = []*dbAddress{}
		customersMap[p.UserID] = &p
		customerIDs = append(customerIDs, p.UserID)
	}

	if err = customerRows.Err(); err != nil {
		return nil, err
	}

	if len(customerIDs) == 0 {
		return []*dbCustomer{}, nil
	}

	placeholders := make([]string, len(customerIDs))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT
			customer_id,
			address_name,
			sub_district,
			district,
			province,
			postal_code
		FROM addresses
		WHERE customer_id IN (%s)
	`, strings.Join(placeholders, ","))

	args := make([]any, len(customerIDs))
	for i, id := range customerIDs {
		args[i] = id
	}

	addressRows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer addressRows.Close()

	for addressRows.Next() {

		var (
			customerID string
			addr       dbAddress
		)

		if err := addressRows.Scan(
			customerID, &addr.AddressName, &addr.SubDistrict,
			&addr.District, &addr.Province, &addr.PostalCode,
		); err != nil {
			return nil, err
		}

		if customer, ok := customersMap[customerID]; ok {
			customer.Addresses = append(customer.Addresses, &addr)
		} else {
			slog.Warn("Address found for unknown customer", "customerID", customerID)
		}
	}
	if err = addressRows.Err(); err != nil {
		return nil, err
	}

	customers := make([]*dbCustomer, 0, len(customersMap))
	for _, id := range customerIDs {
		customers = append(customers, customersMap[id])
	}

	return customers, nil
}

// Customer returns the customer.
func (s *customerStorage) customer(ctx context.Context, customerID string) (*dbCustomer, error) {
	var customer dbCustomer
	if err := s.db.QueryRowContext(ctx, `
		SELECT 
			customer_id,username,bio,facebook,instagram,line,create_time,update_time
		FROM customers 
		WHERE customer_id = $1`,
		customerID,
	).Scan(
		&customer.UserID,
		&customer.Username,
		&customer.Bio,
		&customer.Social.Facebook,
		&customer.Social.Instagram,
		&customer.Social.Line,
		&customer.CreateTime,
		&customer.UpdateTime,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}

	addressRows, err := s.db.QueryContext(ctx, `
		SELECT
			address_name,
			sub_district,
			district,
			province,
			postal_code
		FROM
			addresses
		WHERE customer_id = $1`,
		customerID,
	)
	if err != nil {
		return nil, err
	}
	defer addressRows.Close()

	for addressRows.Next() {
		var addr dbAddress
		if err := addressRows.Scan(
			&addr.AddressName,
			&addr.SubDistrict,
			&addr.District,
			&addr.Province,
			&addr.PostalCode,
		); err != nil {
			return nil, err
		}
		customer.Addresses = append(customer.Addresses, &addr)
	}

	if err = addressRows.Err(); err != nil {
		return nil, err
	}

	return &customer, nil
}

// Create inserts new customer with empty fields. it intends to create
// column before update fields.
func (s *customerStorage) create(ctx context.Context, newCustomer *newCustomer) (string, error) {

	res := s.db.QueryRowContext(ctx, `
		INSERT INTO customers(
			customer_id,
			username
		)
		VALUES($1,$2)
		RETURNING customer_id
	`,
		newCustomer.UserID,
		newCustomer.Username,
	)

	var customerID string
	if err := res.Scan(&customerID); err != nil {
		return "", err
	}

	return customerID, nil
}

// CreateAddress inserts new address to customer and return customerID
func (s *customerStorage) createAddress(ctx context.Context, customerID string, newAddress *dbAddress) (string, error) {

	row := s.db.QueryRowContext(ctx, `
		INSERT INTO addresses(
			customer_id,
			address_name,
			sub_district,
			district,
			province,
			postal_code
		)
		VALUES($1,$2,$3,$4,$5,$6)
		RETURNING customer_id
	`,
		customerID,
		newAddress.AddressName.String,
		newAddress.SubDistrict.String,
		newAddress.District.String,
		newAddress.Province.String,
		newAddress.PostalCode.String,
	)

	var ID string
	if err := row.Scan(&ID); err != nil {
		return "", err
	}

	return ID, nil

}

// Update updates the specified fields in a customer. Only non-empty fields
// in the `update` parameter will overwrite existing values in the database.
// If a field in `update` is an empty string or `NULL`, the corresponding field
// in the database will retain its current value.
//
// COALESCE(arg1, arg2, ...) returns the first non-null value. `arg1` is a bind variable,
// meaning if `arg1` is non-null, it is used as the new value for the update.
// Otherwise, the current database value (`arg2`) is used.
//
// When the application passes a primitive type with a default value, such as false, "",
// or 0 via a bind variable, COALESCE will not treat this as a non-null value. This means
// that fields will be updated even if they are empty or default values.
// To address this, we use NULLIF() to compare against the empty value, returning NULL
// if the value matches the empty or default case.
//
// NULLIF(expr1, expr2) returns NULL if `expr1` and `expr2` are equal.
func (s *customerStorage) update(ctx context.Context, customerID string, update *dbCustomer) (string, error) {

	row := s.db.QueryRowContext(ctx, `
		UPDATE 
			customers
		SET
		    username = COALESCE(NULLIF($2, ''), username),
		    bio = COALESCE(NULLIF($3,''), bio),
		    facebook = COALESCE(NULLIF($4,''), facebook),
		    instagram = COALESCE(NULLIF($5,''), instagram),
		    line = COALESCE(NULLIF($6,''), line),
			update_time = NOW()
		WHERE 
			customer_id = $1
		RETURNING customer_id;
	`,
		customerID,
		update.Username,
		update.Bio,
		update.Social.Facebook,
		update.Social.Instagram,
		update.Social.Line,
	)

	var updatedID string
	if err := row.Scan(&updatedID); err != nil {
		return "", err
	}

	return updatedID, nil

}

// remove the customer.
func (s *customerStorage) remove(ctx context.Context, customerID string) error {

	if _, err := s.db.ExecContext(ctx, `DELETE FROM customers WHERE customer_id=$1`, customerID); err != nil {
		return err
	}
	return nil

}

func (s *customerStorage) countAddress(ctx context.Context, customerID string) (int, error) {

	row := s.db.QueryRowContext(ctx, `SELECT COUNT(customer_id) FROM addresses WHERE customer_id=$1`, customerID)

	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}

	return n, nil
}
