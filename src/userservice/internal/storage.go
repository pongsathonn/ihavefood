package internal

import (
	"context"
	"database/sql"
)

type UserStorage interface {

	// Create creates new user profile.
	Create(ctx context.Context, newProfile *dbProfile) (*dbProfile, error)

	// Profiles returns a list of user profiles.
	Profiles(ctx context.Context) ([]*dbProfile, error)

	// Profile returns the user profile.
	Profile(ctx context.Context, userId string) (*dbProfile, error)

	// Delete deletes the user profile.
	Delete(ctx context.Context, userId string) error
}

type userStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) UserStorage {
	return &userStorage{db: db}
}

func (r *userStorage) Create(ctx context.Context, newProfile *dbProfile) (*dbProfile, error) {
	res := r.db.QueryRowContext(ctx, `
		INSERT INTO profile(
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
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW())
	`,
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
	var profile dbProfile
	if err := res.Scan(&profile); err != nil {
		return nil, err
	}

	return &profile, nil

}

func (r *userStorage) Profiles(ctx context.Context) ([]*dbProfile, error) {

	rows, err := r.db.QueryContext(ctx, `
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
			profile.create_time
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

func (r *userStorage) Profile(ctx context.Context, userID string) (*dbProfile, error) {

	row := r.db.QueryRowContext(ctx, `
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
			profile.create_time
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

func (r *userStorage) Delete(ctx context.Context, userID string) error {

	_, err := r.db.ExecContext(ctx, `DELETE FROM profile WHERE id=$1`, userID)
	if err != nil {
		return err
	}

	return nil
}
