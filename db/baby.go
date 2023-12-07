package db

import (
	"NourishNestApp/model"
	"log"
	"time"

	"github.com/google/uuid"
)

func InsertBaby(b *model.Baby) (err error) {
	rawStmt := `
  INSERT INTO babies (id, user_id, first_name, last_name, date_of_birth, gender, created_at, updated_at)
  VALUES
  (?, ?, ?, ?, ?, ?, ?, ?)
  `

	stmt, err := Db.Prepare(rawStmt)
	if err != nil {
		log.Printf("Failed to parse insert baby SQL statement")
		return err
	}
	defer stmt.Close()

	id := uuid.New().String()
	_, err = stmt.Exec(
		id,
		b.User.Id,
		b.FirstName,
		b.LastName,
		b.DateOfBirth,
		b.Gender,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		log.Printf("Failed to create new baby: %s", err)
		return err
	}

	log.Printf("Created new baby: " + b.Id)
	return
}

func GetBabiesByUser(user *model.User) (babies []model.Baby, err error) {
	babiesMap := make(map[string]*model.Baby)
	var babyKeys []string

	rows, err := Db.Query("SELECT id, first_name, last_name, date_of_birth, gender, created_at, updated_at FROM babies WHERE user_id = ? ORDER BY created_at DESC", user.Id)
	if err != nil {
		log.Printf("Failed to get babies by user id: %s", err)
		return babies, err
	}

	defer rows.Close()

	for rows.Next() {
		b := model.Baby{
			Id:          "",
			Gender:      "",
			FirstName:   "",
			LastName:    "",
			DateOfBirth: time.Time{},
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
			User:        *user,
		}

		err = rows.Scan(
			&b.Id,
			&b.FirstName,
			&b.LastName,
			&b.DateOfBirth,
			&b.Gender,
			&b.CreatedAt,
			&b.UpdatedAt,
		)

		if err != nil {
			panic(err.Error())
		}
		babyKeys = append(babyKeys, b.Id)
		babiesMap[b.Id] = &b
	}

	for _, babyId := range babyKeys {
		baby := *babiesMap[babyId]
		babies = append(babies, baby)
	}

	return babies, nil
}
