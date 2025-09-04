package internal

import (
	"context"
	"database/sql"
	"github.com/lib/pq"
	"log/slog"
)

func NewCustomerStorage(db *sql.DB) *customerStorage {
	return &customerStorage{db: db}
}

type customerStorage struct {
	db *sql.DB
}

func (s *customerStorage) listCustomers(ctx context.Context) ([]*dbCustomer, error) {
	customerRows, err := s.db.QueryContext(ctx, `
		SELECT 
			customer_id, username, bio, facebook, instagram, line, create_time, update_time
		FROM customers
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
			&p.CustomerID, &p.Username, &p.Bio, &p.Social.Facebook,
			&p.Social.Instagram, &p.Social.Line, &p.CreateTime, &p.UpdateTime,
		)
		if err != nil {
			return nil, err
		}
		p.Addresses = []*dbAddress{}
		customersMap[p.CustomerID] = &p
		customerIDs = append(customerIDs, p.CustomerID)
	}
	if err = customerRows.Err(); err != nil {
		return nil, err
	}

	if len(customerIDs) == 0 {
		return []*dbCustomer{}, nil
	}

	query := `
		SELECT
			customer_id,
			address_name,
			sub_district,
			district,
			province,
			postal_code
		FROM addresses
		WHERE customer_id IN =  ANY($1)
	`

	addressRows, err := s.db.QueryContext(ctx, query, pq.Array(customerIDs))
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
			&customerID,
			&addr.AddressName,
			&addr.SubDistrict,
			&addr.District,
			&addr.Province,
			&addr.PostalCode,
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

func (s *customerStorage) getCustomer(ctx context.Context, customerID string) (*dbCustomer, error) {
	var customer dbCustomer
	if err := s.db.QueryRowContext(ctx, `
		SELECT 
			customer_id,username,bio,facebook,instagram,line,create_time,update_time
		FROM customers 
		WHERE customer_id = $1`,
		customerID,
	).Scan(
		&customer.CustomerID,
		&customer.Username,
		&customer.Bio,
		&customer.Social.Facebook,
		&customer.Social.Instagram,
		&customer.Social.Line,
		&customer.CreateTime,
		&customer.UpdateTime,
	); err != nil {
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

func (s *customerStorage) getAddress(ctx context.Context, customerID, addressID string) (*dbAddress, error) {
	var addr dbAddress
	row := s.db.QueryRowContext(ctx, `
        SELECT 
            address_id,
            address_name,
            sub_district,
            district,
            province,
            postal_code
        FROM addresses
        WHERE customer_id = $1 AND address_id = $2
    `, customerID, addressID)

	if err := row.Scan(
		&addr.AddressID,
		&addr.AddressName,
		&addr.SubDistrict,
		&addr.District,
		&addr.Province,
		&addr.PostalCode,
	); err != nil {
		return nil, err
	}

	return &addr, nil
}

func (s *customerStorage) create(ctx context.Context, newCustomer *newCustomer) (string, error) {

	res := s.db.QueryRowContext(ctx, `
		INSERT INTO customers(
			customer_id,
			username
		)
		VALUES($1,$2)
		RETURNING customer_id
	`,
		newCustomer.CustomerID,
		newCustomer.Username,
	)

	var customerID string
	if err := res.Scan(&customerID); err != nil {
		return "", err
	}

	return customerID, nil
}

func (s *customerStorage) createAddress(ctx context.Context, customerID string, newAddress *dbAddress) (string, error) {

	row := s.db.QueryRowContext(ctx, `
    INSERT INTO addresses (
        customer_id,
        address_name,
        sub_district,
        district,
        province,
        postal_code
    )
    VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING address_id
	`,
		customerID,
		newAddress.AddressName.String,
		newAddress.SubDistrict.String,
		newAddress.District.String,
		newAddress.Province.String,
		newAddress.PostalCode.String,
	)

	var addressID string
	if err := row.Scan(&addressID); err != nil {
		return "", err
	}

	return addressID, nil
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

func (s *customerStorage) remove(ctx context.Context, customerID string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM customers WHERE customer_id=$1`, customerID); err != nil {
		return err
	}
	return nil
}
