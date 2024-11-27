package internal

import (
	"context"
	"database/sql"
)

type ProfileStorage interface {

	// Profiles returns a list of user profiles.
	Profiles(ctx context.Context) ([]*dbProfile, error)

	// Profile returns the user profile.
	Profile(ctx context.Context, userID string) (*dbProfile, error)

	// Create creates new user profile with empty fields. it intends to create
	// column before update fields.
	Create(ctx context.Context, newProfile *newProfile) (string, error)

	// Update user profile.
	Update(ctx context.Context, userID string, update *dbProfile) (string, error)

	// Delete deletes the user profile.
	Delete(ctx context.Context, userID string) error
}

type profileStorage struct {
	db *sql.DB
}

func NewProfileStorage(db *sql.DB) ProfileStorage {
	return &profileStorage{db: db}
}

func (s *profileStorage) Profiles(ctx context.Context) ([]*dbProfile, error) {

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id,
			username,
			bio,
			facebook,
			instagram,
			line,
			address_name,
			sub_district,
			district,
			province,
			postal_code,
			create_time,
			update_time
		FROM
			profile
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*dbProfile
	for rows.Next() {
		profile := dbProfile{
			Social:  &dbSocial{},
			Address: &dbAddress{},
		}
		err := rows.Scan(
			&profile.UserID,
			&profile.Username,
			&profile.Bio,
			&profile.Social.Facebook,
			&profile.Social.Instragram,
			&profile.Social.Line,
			&profile.Address.AddressName,
			&profile.Address.SubDistrict,
			&profile.Address.District,
			&profile.Address.Province,
			&profile.Address.PostalCode,
			&profile.CreateTime,
			&profile.UpdateTime,
		)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, &profile)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}

func (s *profileStorage) Profile(ctx context.Context, userID string) (*dbProfile, error) {

	row := s.db.QueryRowContext(ctx, `
		SELECT
			id,
			username,
			bio,
			facebook,
			instagram,
			line,
			address_name,
			sub_district,
			district,
			province,
			postal_code,
			create_time,
			update_time
		FROM
			profile
		WHERE
			id = $1;
	`, userID)

	// Dereferencing uninitilized embedded struct in GO will cause
	// a runtime error. To fix this  initialize the embedded structs
	// before using them.
	profile := dbProfile{
		Social:  &dbSocial{},
		Address: &dbAddress{},
	}

	err := row.Scan(
		&profile.UserID,
		&profile.Username,
		&profile.Bio,
		&profile.Social.Facebook,
		&profile.Social.Instragram,
		&profile.Social.Line,
		&profile.Address.AddressName,
		&profile.Address.SubDistrict,
		&profile.Address.District,
		&profile.Address.Province,
		&profile.Address.PostalCode,
		&profile.CreateTime,
		&profile.UpdateTime,
	)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (s *profileStorage) Create(ctx context.Context, newProfile *newProfile) (string, error) {

	res := s.db.QueryRowContext(ctx, `
		INSERT INTO profile(
			id,
			username
		)
		VALUES($1,$2)
		RETURNING id
	`,
		newProfile.UserID,
		newProfile.Username,
	)

	var userID string
	if err := res.Scan(&userID); err != nil {
		return "", err
	}

	return userID, nil
}

// Update updates the specified fields in a user profile. Only non-empty fields
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
func (s *profileStorage) Update(ctx context.Context, userID string, update *dbProfile) (string, error) {

	row := s.db.QueryRowContext(ctx, `
		UPDATE 
			profile
		SET
		    username = COALESCE(NULLIF($2, ''), username),
		    bio = COALESCE(NULLIF($4,''), bio),

		    facebook = COALESCE(NULLIF($5,''), facebook),
		    instagram = COALESCE(NULLIF($6,''), instagram),
		    line = COALESCE(NULLIF($7,''), line),

		    address_name = COALESCE(NULLIF($8,''), address_name),
		    sub_district = COALESCE(NULLIF($9,''), sub_district),
		    district = COALESCE(NULLIF($10,''), district),
		    province = COALESCE(NULLIF($11,''), province),
		    postal_code = COALESCE(NULLIF($12,''), postal_code),
			update_time = NOW()

		WHERE id = $1
		RETURNING id;
	`,
		userID,
		update.Username,
		update.Bio,

		update.Social.Facebook,
		update.Social.Instragram,
		update.Social.Line,

		update.Address.AddressName.String,
		update.Address.SubDistrict.String,
		update.Address.District.String,
		update.Address.Province.String,
		update.Address.PostalCode.String,
	)

	var updatedID string
	if err := row.Scan(&updatedID); err != nil {
		return "", err
	}

	return updatedID, nil

}

func (s *profileStorage) Delete(ctx context.Context, userID string) error {

	if _, err := s.db.ExecContext(ctx, `DELETE FROM profile WHERE id=$1`, userID); err != nil {
		return err
	}
	return nil

}
