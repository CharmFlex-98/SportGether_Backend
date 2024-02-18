package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"
)

type UserProfileDao struct {
	db *sql.DB
}

type UserProfileDetail struct {
	PreferredName       *string   `json:"preferredName"`
	BirthDate           *string   `json:"birthDate"`
	Signature           *string   `json:"signature"`
	Memo                *string   `json:"memo"`
	JoinTime            time.Time `json:"joinTime"`
	ProfileIconUrl      *string   `json:"profileIconUrl"`
	ProfileIconPublicId *string   `json:"profileIconPublicId"`
	Gender              *string   `json:"gender"`
}

func (profileDao UserProfileDao) UserIsOnboarded(userId int64) (bool, error) {
	query := `SELECT up.status FROM sportgether_schema.user_profile up WHERE up.user_id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var onboardStatus string
	err := profileDao.db.QueryRowContext(ctx, query, userId).Scan(&onboardStatus)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func (profileDao UserProfileDao) GetProfileDetail(userId int64) (*UserProfileDetail, error) {
	query := `
	SELECT 
	    preferred_name, 
	    gender, 
	    birth_date, 
	    join_date, 
	    profile_icon_url, 
	    signature, 
	    memo 
	FROM sportgether_schema.user_profile up
	WHERE up.user_id = $1
`

	userProfileDetail := &UserProfileDetail{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := profileDao.db.QueryRowContext(ctx, query, userId).Scan(
		&userProfileDetail.PreferredName,
		&userProfileDetail.Gender,
		&userProfileDetail.BirthDate,
		&userProfileDetail.JoinTime,
		&userProfileDetail.ProfileIconUrl,
		&userProfileDetail.Signature,
		&userProfileDetail.Memo,
	)
	if err != nil {
		return nil, err
	}

	return userProfileDetail, nil
}

func (profileDao UserProfileDao) OnboardUser(userId int64, preferredName string, birthDate string, gender string) (bool, error) {
	query := `
	INSERT INTO sportgether_schema.user_profile(user_id, preferred_name, birth_date, gender, status)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING status
`
	args := []any{
		userId,
		preferredName,
		birthDate,
		gender,
		"ONBOARDED",
	}

	var status string

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := profileDao.db.QueryRowContext(ctx, query, args...).Scan(&status)
	if err != nil {
		return false, err
	}

	if status == "ONBOARDED" {
		return true, nil
	}

	return false, nil
}

func (profileDao UserProfileDao) GetUserProfileIconUrl(userId int64) (*string, *string, error) {
	query := `
		SELECT profile_icon_url, profile_icon_public_id from sportgether_schema.user_profile WHERE user_id=$1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var profileIconUrl *string
	var profileIconPublicId *string
	err := profileDao.db.QueryRowContext(ctx, query, userId).Scan(&profileIconUrl, &profileIconPublicId)
	if err != nil {
		return nil, nil, err
	}

	return profileIconUrl, profileIconPublicId, nil
}

func (profileDao UserProfileDao) UpdateUserProfile(userId int64, detail UserProfileDetail) error {
	columnsMap := map[string]any{
		"preferred_name":         detail.PreferredName,
		"birth_date":             detail.BirthDate,
		"signature":              detail.Signature,
		"gender":                 detail.Gender,
		"profile_icon_url":       detail.ProfileIconUrl,
		"profile_icon_public_id": detail.ProfileIconPublicId,
		"memo":                   detail.Memo,
	}
	setQuery, values := buildColumnsToUpdate(columnsMap)
	whereClause := fmt.Sprintf("WHERE up.user_id = $%d", len(values)+1)
	values = append(values, userId)

	query := fmt.Sprintf(`
	UPDATE sportgether_schema.user_profile up
	%s
	%s
`, setQuery, whereClause)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := profileDao.db.ExecContext(ctx, query, values...)
	if err != nil {
		return err
	}

	return nil
}

func buildColumnsToUpdate(columnNames map[string]any) (string, []any) {
	setColumn := "SET "
	var values []any

	count := 0
	for columnName, value := range columnNames {
		if !reflect.ValueOf(value).IsNil() {
			if count == 0 {
				setColumn += fmt.Sprintf("%s = $%d", columnName, len(values)+1)
			} else {
				setColumn += ", " + fmt.Sprintf("%s = $%d", columnName, len(values)+1)
			}
			values = append(values, value)
			count++
		}
	}

	return setColumn, values
}
