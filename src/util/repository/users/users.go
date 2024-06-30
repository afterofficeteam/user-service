package users

import (
	"database/sql"
	"fmt"
	"strings"
	"user-service/src/util/helper"
	"user-service/src/util/repository/model/users"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *store {
	return &store{
		db: db,
	}
}

func (s *store) RegisterUser(bReq users.Users) (*uuid.UUID, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	var userID uuid.UUID
	queryCreate := `
		INSERT INTO users(
			email,
		    username,
			role,
 			address,
 			category_preferences,
			created_at
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			now()
		) RETURNING id
	`

	if err := tx.QueryRow(
		queryCreate,
		bReq.Email,
		bReq.Username,
		bReq.Role,
		bReq.Address,
		pq.Array(bReq.CategoryPreferences),
	).Scan(&userID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &userID, nil
}

func (s *store) GetUserDetails(bReq users.Users) (*users.Users, error) {
	querySelect := `
		SELECT
			*
		FROM
		    users
	`

	var queryConditions []string
	if bReq.Email != "" {
		queryConditions = append(queryConditions, fmt.Sprintf("email = '%s'", bReq.Email))
	}

	if bReq.Id != uuid.Nil {
		queryConditions = append(queryConditions, fmt.Sprintf("id = '%v'", bReq.Id))
	}

	if len(queryConditions) > 0 {
		querySelect += " WHERE " + strings.Join(queryConditions, " AND ")
	}

	querySelect += `
		ORDER BY created_at DESC limit 1
	`

	var response users.Users
	rows, err := s.db.Query(querySelect)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&response.Id,
			&response.Email,
			&response.Username,
			&response.Role,
			&response.Address,
			pq.Array(&response.CategoryPreferences),
			&response.CreatedAt,
			&response.UpdatedAt,
			&response.DeletedAt,
		); err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("no partner found")
			}
			return nil, fmt.Errorf("failed to fetch user data")
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterate over user: %v", err)
	}

	return &response, nil
}

func (s *store) GetUsers(bReq users.RequestUsers) (*[]users.Users, int, error) {
	querySelect := `
		SELECT
			*
		FROM
		    users
	`

	var queryConditions []string
	if bReq.UserId != uuid.Nil {
		queryConditions = append(queryConditions, fmt.Sprintf("id = '%v'", bReq.UserId))
	}

	if bReq.Email != "" {
		queryConditions = append(queryConditions, fmt.Sprintf("email = '%s'", bReq.Email))
	}

	if bReq.Search != "" {
		searchTerm := fmt.Sprintf("%%%s%%", bReq.Search)
		queryConditions = append(queryConditions, fmt.Sprintf("(email LIKE '%s' OR username LIKE '%s' OR role LIKE '%s')", searchTerm, searchTerm, searchTerm))
	}

	if bReq.Role != "" {
		queryConditions = append(queryConditions, fmt.Sprintf("role = '%s'", bReq.Role))
	}

	if len(queryConditions) > 0 {
		querySelect += "WHERE " + strings.Join(queryConditions, " AND ")
	}

	querySelect += `
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	offset := (bReq.Page - 1) * bReq.Limit
	rows, err := s.db.Query(querySelect, bReq.Limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var usersData []users.Users
	for rows.Next() {
		var user users.Users
		if err := rows.Scan(
			&user.Id,
			&user.Email,
			&user.Username,
			&user.Role,
			&user.Address,
			pq.Array(&user.CategoryPreferences),
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan rows: %v", err)
		}
		usersData = append(usersData, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate rows: %v", err)
	}

	totalData := len(usersData)

	return &usersData, totalData, nil
}
func (s *store) UpdateUser(id uuid.UUID, bReq users.Users) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	queryLock := `
        SELECT 1
        FROM users
        WHERE id = $1
        FOR UPDATE
    `
	if _, err := tx.Exec(queryLock, id); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to lock user row: %w", err)
	}

	queryUpdate := `
        UPDATE users
        SET
            email = $1,
            role = $2,
            address = $3,
            category_preferences = $4,
            updated_at = $5
        WHERE
            id = $6
    `

	timeNow, err := helper.TimeNow()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(
		queryUpdate,
		bReq.Email,
		bReq.Role,
		bReq.Address,
		pq.Array(bReq.CategoryPreferences),
		&timeNow,
		id,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
