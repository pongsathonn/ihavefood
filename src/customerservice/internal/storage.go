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
		SELECT customer_id, username, facebook, instagram, line, create_time, update_time
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
			&p.CustomerID, &p.Username, &p.Social.Facebook,
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
			address_id,
			address_name,
			sub_district,
			district,
			province,
			postal_code
		FROM addresses
		WHERE customer_id =  ANY($1)
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
			&addr.AddressID,
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
			customer_id,username,email,facebook,instagram,line,create_time,update_time
		FROM customers 
		WHERE customer_id = $1`,
		customerID,
	).Scan(
		&customer.CustomerID,
		&customer.Username,
		&customer.Email,
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
			address_id,
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
			&addr.AddressID,
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

func (s *customerStorage) create(ctx context.Context, newCustomer *dbNewCustomer) (string, error) {

	res := s.db.QueryRowContext(ctx, `
		INSERT INTO customers(
			customer_id,
			username,
			email
		)
		VALUES($1,$2,$3)
		RETURNING customer_id
	`,
		newCustomer.CustomerID,
		newCustomer.Username,
		newCustomer.Email,
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
		newAddress.AddressName,
		newAddress.SubDistrict,
		newAddress.District,
		newAddress.Province,
		newAddress.PostalCode,
	)

	var addressID string
	if err := row.Scan(&addressID); err != nil {
		return "", err
	}

	return addressID, nil
}

func (s *customerStorage) updateCustomerInfo(ctx context.Context, customerID, username, phone string) (string, error) {
	row := s.db.QueryRowContext(ctx, `
    UPDATE customers
    SET
      username = COALESCE(NULLIF($2, ''), username),
	  phone    = COALESCE(NULLIF($3, ''), phone),
      update_time = NOW()
    WHERE customer_id = $1
    RETURNING customer_id;
  `,
		customerID,
		username,
	)

	var id string
	if err := row.Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

func (s *customerStorage) updateCustomerSocial(ctx context.Context, customerID string, social *dbSocial) (string, error) {

	row := s.db.QueryRowContext(ctx, `
    UPDATE customers
    SET
      facebook  = COALESCE(NULLIF($2,''), facebook),
      instagram = COALESCE(NULLIF($3,''), instagram),
      line      = COALESCE(NULLIF($4,''), line),
      update_time = NOW()
    WHERE customer_id = $1
    RETURNING customer_id;
  `,
		customerID,
		social.Facebook,
		social.Instagram,
		social.Line,
	)

	var id string
	if err := row.Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

func (s *customerStorage) updateCustomerAddress(ctx context.Context, customerID, addressID string, addr *dbAddress) (string, error) {

	row := s.db.QueryRowContext(ctx, `
    UPDATE addresses
    SET
      address_name = COALESCE(NULLIF($3,''), address_name),
      sub_district = COALESCE(NULLIF($4,''), sub_district),
      district     = COALESCE(NULLIF($5,''), district),
      province     = COALESCE(NULLIF($6,''), province),
      postal_code  = COALESCE(NULLIF($7,''), postal_code)
    WHERE address_id = $2
      AND customer_id = $1
    RETURNING address_id;
  `,
		customerID,
		addressID,
		addr.AddressName,
		addr.SubDistrict,
		addr.District,
		addr.Province,
		addr.PostalCode,
	)

	var id string
	if err := row.Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

func (s *customerStorage) remove(ctx context.Context, customerID string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM customers WHERE customer_id=$1`, customerID); err != nil {
		return err
	}
	return nil
}
