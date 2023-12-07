package model

import "time"

type (
	User struct {
		CreatedAt time.Time
		UpdatedAt time.Time
		Id        string
		Token     string
		Email     string
		Name      string
	}
)
