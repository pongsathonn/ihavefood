package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

var ErrProfileNotFound = errors.New("profile not found")

func NewProfileStorage(db *sql.DB) *profileStorage {
	return &profileStorage{db: db}
}

type profileStorage struct {
	db *sql.DB
}

// Profiles returns a list of user profiles.
func (s *profileStorage) profiles(ctx context.Context) ([]*dbProfile, error) {
	profileRows, err := s.db.QueryContext(ctx, `
		SELECT
			user_id, username, bio, facebook, instagram, line, create_time, update_time
		FROM profiles
		ORDER BY CAST(user_id AS INTEGER)
	`)
	if err != nil {
		return nil, err
	}
	defer profileRows.Close()

	profilesMap := make(map[string]*dbProfile)
	var userIDs []string

	for profileRows.Next() {
		var p dbProfile
		err := profileRows.Scan(
			&p.UserID, &p.Username, &p.Bio, &p.Social.Facebook,
			&p.Social.Instagram, &p.Social.Line, &p.CreateTime, &p.UpdateTime,
		)
		if err != nil {
			return nil, err
		}
		p.Addresses = []*dbAddress{}
		profilesMap[p.UserID] = &p
		userIDs = append(userIDs, p.UserID)
	}

	if err = profileRows.Err(); err != nil {
		return nil, err
	}

	if len(userIDs) == 0 {
		return []*dbProfile{}, nil
	}

	placeholders := make([]string, len(userIDs))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT
			user_id,
			address_name,
			sub_district,
			district,
			province,
			postal_code
		FROM addresses
		WHERE user_id IN (%s)
	`, strings.Join(placeholders, ","))

	args := make([]any, len(userIDs))
	for i, id := range userIDs {
		args[i] = id
	}

	addressRows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer addressRows.Close()

	for addressRows.Next() {

		var (
			userID string
			addr   dbAddress
		)

		if err := addressRows.Scan(
			userID, &addr.AddressName, &addr.SubDistrict,
			&addr.District, &addr.Province, &addr.PostalCode,
		); err != nil {
			return nil, err
		}

		if profile, ok := profilesMap[userID]; ok {
			profile.Addresses = append(profile.Addresses, &addr)
		} else {
			slog.Warn("Address found for unknown user", "userID", userID)
		}
	}
	if err = addressRows.Err(); err != nil {
		return nil, err
	}

	profiles := make([]*dbProfile, 0, len(profilesMap))
	for _, id := range userIDs {
		profiles = append(profiles, profilesMap[id])
	}

	return profiles, nil
}

// Profile returns the user profile.
func (s *profileStorage) profile(ctx context.Context, userID string) (*dbProfile, error) {
	var profile dbProfile
	if err := s.db.QueryRowContext(ctx, `
		SELECT 
			user_id,username,bio,facebook,instagram,line,create_time,update_time
		FROM profiles 
		WHERE user_id = $1`,
		userID,
	).Scan(
		&profile.UserID,
		&profile.Username,
		&profile.Bio,
		&profile.Social.Facebook,
		&profile.Social.Instagram,
		&profile.Social.Line,
		&profile.CreateTime,
		&profile.UpdateTime,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProfileNotFound
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
		WHERE user_id = $1`,
		userID,
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
		profile.Addresses = append(profile.Addresses, &addr)
	}

	if err = addressRows.Err(); err != nil {
		return nil, err
	}

	return &profile, nil
}

// Create inserts new user profile with empty fields. it intends to create
// column before update fields.
func (s *profileStorage) create(ctx context.Context, newProfile *newProfile) (string, error) {

	res := s.db.QueryRowContext(ctx, `
		INSERT INTO profiles(
			user_id,
			username
		)
		VALUES($1,$2)
		RETURNING user_id
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

// CreateAddress inserts new address to user profile and return userID
func (s *profileStorage) createAddress(ctx context.Context, userID string, newAddress *dbAddress) (string, error) {

	row := s.db.QueryRowContext(ctx, `
		INSERT INTO addresses(
			user_id,
			address_name,
			sub_district,
			district,
			province,
			postal_code
		)
		VALUES($1,$2,$3,$4,$5,$6)
		RETURNING user_id
	`,
		userID,
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
			profiles
		SET
		    username = COALESCE(NULLIF($2, ''), username),
		    bio = COALESCE(NULLIF($3,''), bio),
		    facebook = COALESCE(NULLIF($4,''), facebook),
		    instagram = COALESCE(NULLIF($5,''), instagram),
		    line = COALESCE(NULLIF($6,''), line),
			update_time = NOW()
		WHERE 
			user_id = $1
		RETURNING user_id;
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

	if _, err := s.db.ExecContext(ctx, `DELETE FROM profiles WHERE user_id=$1`, userID); err != nil {
		return err
	}
	return nil

}

func (s *profileStorage) countAddress(ctx context.Context, userID string) (int, error) {

	row := s.db.QueryRowContext(ctx, `SELECT COUNT(user_id) FROM addresses WHERE user_id=$1`, userID)

	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}

	return n, nil
}
