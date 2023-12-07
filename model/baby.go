package model

import "time"

type (
	Baby struct {
		Id          string
		Gender      string
		FirstName   string
		LastName    string
		DateOfBirth time.Time
		CreatedAt   time.Time
		UpdatedAt   time.Time
		User        User
	}
	NewBaby struct {
		Gender      string `form:"gender"`
		FirstName   string `form:"first_name"`
		LastName    string `form:"last_name"`
		DateOfBirth string `form:"date_of_birth"`
		User        User
	}
)
