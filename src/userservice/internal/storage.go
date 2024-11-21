package internal

import (
	"context"
	"database/sql"
)

type UserStorage interface {

	// Profiles returns a list of user profiles.
	Profiles(ctx context.Context) ([]*dbProfile, error)

	// Profile returns the user profile.
	Profile(ctx context.Context, userID string) (*dbProfile, error)

	// Create creates new user profile.
	Create(ctx context.Context, newProfile *dbProfile) (string, error)

	// Update user profile.
	Update(ctx context.Context, userID string, update *dbProfile) (string, error)

	// Delete deletes the user profile.
	Delete(ctx context.Context, userID string) error
}

type userStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) UserStorage {
	return &userStorage{db: db}
}

func (s *userStorage) Profiles(ctx context.Context) ([]*dbProfile, error) {

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			profile.id,
			profile.username,
			profile.picture,
			profile.bio,
			profile.facebook,
			profile.instagram,
			profile.line,
			profile.address_name,
			profile.sub_district,
			profile.district,
			profile.province,
			profile.postal_code,
			profile.create_time,
		FROM
			profile
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*dbProfile
	for rows.Next() {
		var profile dbProfile
		err := rows.Scan(
			&profile.UserID,
			&profile.Username,
			&profile.Picture,
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

func (s *userStorage) Profile(ctx context.Context, userID string) (*dbProfile, error) {

	row := s.db.QueryRowContext(ctx, `
		SELECT
			profile.id,
			profile.username,
			profile.picture,
			profile.bio,
			profile.facebook,
			profile.instagram,
			profile.line,
			profile.address_name,
			profile.sub_district,
			profile.district,
			profile.province,
			profile.postal_code,
			profile.create_time,
		FROM
			profile
		WHERE
			profile.id = $1;
	`, userID)

	var profile dbProfile
	err := row.Scan(
		&profile.UserID,
		&profile.Username,
		&profile.Picture,
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
	)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (s *userStorage) Create(ctx context.Context, newProfile *dbProfile) (string, error) {

	res := s.db.QueryRowContext(ctx, `
		INSERT INTO profile(
			id,
			username,
			picture,
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
		)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW())
		RETURNING id
	`,
		newProfile.UserID,
		newProfile.Username,
		newProfile.Picture,
		newProfile.Bio,
		newProfile.Social.Facebook,
		newProfile.Social.Instragram,
		newProfile.Social.Line,
		newProfile.Address.AddressName,
		newProfile.Address.SubDistrict,
		newProfile.Address.District,
		newProfile.Address.Province,
		newProfile.Address.PostalCode,
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
func (s *userStorage) Update(ctx context.Context, userID string, update *dbProfile) (string, error) {

	row := s.db.QueryRowContext(ctx, `
		UPDATE 
			profile
		SET
		    username = COALESCE(NULLIF($2, ''), username),
		    picture = COALESCE($3, picture),
		    bio = COALESCE(NULLIF($4,''), bio),

		    facebook = COALESCE(NULLIF($5,''), facebook),
		    instagram = COALESCE(NULLIF($6,''), instagram),
		    line = COALESCE(NULLIF($7,''), line),

		    address_name = COALESCE(NULLIF($8,''), address_name),
		    sub_district = COALESCE(NULLIF($9,''), sub_district),
		    district = COALESCE(NULLIF($10,''), district),
		    province = COALESCE(NULLIF($11,''), province),
		    postal_code = COALESCE(NULLIF($12,''), postal_code)

		WHERE id = $1
		RETURNING id;
	`,
		userID,
		update.Username,
		update.Picture,
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

func (s *userStorage) Delete(ctx context.Context, userID string) error {

	if _, err := s.db.ExecContext(ctx, `DELETE FROM profile WHERE id=$1`, userID); err != nil {
		return err
	}
	return nil

}
