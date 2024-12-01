package internal

import (
	"context"
	"database/sql"
)

func NewProfileStorage(db *sql.DB) *profileStorage {
	return &profileStorage{db: db}
}

type profileStorage struct {
	db *sql.DB
}

// Profiles returns a list of user profiles.
func (s *profileStorage) profiles(ctx context.Context) ([]*dbProfile, error) {

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			profiles.profile_id,
			profiles.username,
			profiles.bio,
			profiles.facebook,
			profiles.instagram,
			profiles.line,
			profiles.create_time,
			profiles.update_time,
			addresses.address_name,
			addresses.sub_district,
			addresses.district,
			addresses.province,
			addresses.postal_code
		FROM
			profiles
		LEFT JOIN addresses USING (profile_id)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var m = make(map[string]*dbProfile)
	for rows.Next() {
		var (
			p       dbProfile
			social  dbSocial
			address dbAddress
		)

		err := rows.Scan(
			&p.UserID,
			&p.Username,
			&p.Bio,
			&social.Facebook,
			&social.Instagram,
			&social.Line,
			&p.CreateTime,
			&p.UpdateTime,
			&address.AddressName,
			&address.SubDistrict,
			&address.District,
			&address.Province,
			&address.PostalCode,
		)
		if err != nil {
			return nil, err
		}

		if _, exists := m[p.UserID]; !exists {
			// initilize the key first
			m[p.UserID] = &dbProfile{
				UserID:     p.UserID,
				Username:   p.Username,
				Bio:        p.Bio,
				Social:     &social,
				Addresses:  []*dbAddress{},
				CreateTime: p.CreateTime,
				UpdateTime: p.UpdateTime,
			}
		}
		m[p.UserID].Addresses = append(m[p.UserID].Addresses, &address)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	profiles := make([]*dbProfile, 0, len(m))
	for _, profile := range m {
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// Profile returns the user profile.
func (s *profileStorage) profile(ctx context.Context, userID string) (*dbProfile, error) {

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			profiles.profile_id,
			profiles.username,
			profiles.bio,
			profiles.facebook,
			profiles.instagram,
			profiles.line,
			profiles.create_time,
			profiles.update_time,
			addresses.address_name,
			addresses.sub_district,
			addresses.district,
			addresses.province,
			addresses.postal_code
		FROM
			profiles
		LEFT JOIN addresses USING (profile_id)
		WHERE profiles.profile_id = $1;
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profile := make(map[string]*dbProfile)
	for rows.Next() {
		var (
			p       dbProfile
			social  dbSocial
			address dbAddress
		)

		err := rows.Scan(
			&p.UserID,
			&p.Username,
			&p.Bio,
			&social.Facebook,
			&social.Instagram,
			&social.Line,
			&p.CreateTime,
			&p.UpdateTime,
			&address.AddressName,
			&address.SubDistrict,
			&address.District,
			&address.Province,
			&address.PostalCode,
		)
		if err != nil {
			return nil, err
		}

		if _, exists := profile[p.UserID]; !exists {
			// initilize the key first
			profile[p.UserID] = &dbProfile{
				UserID:     p.UserID,
				Username:   p.Username,
				Bio:        p.Bio,
				Social:     &social,
				Addresses:  []*dbAddress{},
				CreateTime: p.CreateTime,
				UpdateTime: p.UpdateTime,
			}
		}
		profile[p.UserID].Addresses = append(profile[p.UserID].Addresses, &address)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return profile[userID], nil
}

// Create creates new user profile with empty fields. it intends to create
// column before update fields.
func (s *profileStorage) create(ctx context.Context, newProfile *newProfile) (string, error) {

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

func (s *profileStorage) updateAddress(ctx context.Context, userID string, newAddr *dbAddress) (string, error) {

	row := s.db.QueryRowContext(ctx, `
		UPDATE 
			profile
		SET
		    address_name = COALESCE(NULLIF($2,''), address_name),
		    sub_district = COALESCE(NULLIF($3,''), sub_district),
		    district = COALESCE(NULLIF($4,''), district),
		    province = COALESCE(NULLIF($5,''), province),
		    postal_code = COALESCE(NULLIF($6,''), postal_code),
			update_time = NOW()
		WHERE id = $1
		RETURNING id;
	`,
		userID,
		newAddr.AddressName.String,
		newAddr.SubDistrict.String,
		newAddr.District.String,
		newAddr.Province.String,
		newAddr.PostalCode.String,
	)

	var updatedID string
	if err := row.Scan(&updatedID); err != nil {
		return "", err
	}

	return updatedID, nil

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
func (s *profileStorage) update(ctx context.Context, userID string, update *dbProfile) (string, error) {

	row := s.db.QueryRowContext(ctx, `
		UPDATE 
			profile
		SET
		    username = COALESCE(NULLIF($2, ''), username),
		    bio = COALESCE(NULLIF($3,''), bio),
		    facebook = COALESCE(NULLIF($4,''), facebook),
		    instagram = COALESCE(NULLIF($5,''), instagram),
		    line = COALESCE(NULLIF($6,''), line),
			update_time = NOW()
		WHERE id = $1
		RETURNING id;
	`,
		userID,
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

// remove deletes the user profile.
func (s *profileStorage) remove(ctx context.Context, userID string) error {

	if _, err := s.db.ExecContext(ctx, `DELETE FROM profile WHERE id=$1`, userID); err != nil {
		return err
	}
	return nil

}
