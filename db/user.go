package db

import (
	"NourishNestApp/logger"
	"NourishNestApp/model"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

func UpsertUser(u model.User) (err error) {
	existingUser, _ := GetUserById(u.Id)

	if existingUser == nil {
		u.CreatedAt = time.Now()
	}

	rawStmt := `
  INSERT OR REPLACE INTO users
  (id, token, email, name, updated_at, created_at)
  VALUES
  (?, ?, ?, ?, ?, ?)
  `

	stmt, err := Db.Prepare(rawStmt)
	if err != nil {
		log.Printf("Failed to parse insert user SQL statement")
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(
		u.Id,
		u.Token,
		u.Email,
		u.Name,
		time.Now(),
		u.CreatedAt,
	)

	if err != nil {
		logger.Log.Warn(fmt.Sprintf("Failed to upsert user: %s", err))
		return err
	}

	if existingUser == nil {
		logger.Log.Info(fmt.Sprintf("Added new user: " + u.Id))
	} else {
		logger.Log.Info(fmt.Sprintf("Updated user: " + u.Id))
	}
	return
}

var SelectUserFieldsSql = "SELECT id, token, email, name, updated_at, created_at FROM users"

func getUserBy(findUserSql string, queryBy any) (user *model.User, err error) {
	var u model.User
	rawStmt := findUserSql

	stmt, err := Db.Prepare(rawStmt)
	if err != nil {
		logger.Log.Warn("failed to parse find user SQL statement")
		return nil, err
	}

	defer stmt.Close()
	if err := stmt.QueryRow(queryBy).Scan(
		&u.Id, &u.Token, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no user found")
		}

		logger.Log.Warn(fmt.Sprintf("failed to convert user from row: %s", err))
		return nil, err
	}

	return &u, nil
}

func GetUserById(id string) (user *model.User, err error) {
	return getUserBy(SelectUserFieldsSql+" WHERE id = ? LIMIT 1", id)
}

func GetUserByEmail(email string) (user *model.User, err error) {
	return getUserBy(SelectUserFieldsSql+" WHERE email = ? LIMIT 1", email)
}

func GetUserByToken(token string) (user *model.User, err error) {
	return getUserBy(SelectUserFieldsSql+" WHERE token = ? LIMIT 1", token)
}
