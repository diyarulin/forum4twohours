package models

import (
	"errors"
)

var ErrNoRecord = errors.New("models: no matching record found")
var ErrInvalidCredentials = errors.New("models: invalid credentials")
var ErrDuplicateEmail = errors.New("email address is already in use")
var ErrIncorrectCurrentPassword = errors.New("password is incorrect")
var ErrDuplicateCategory = errors.New("category already exists")
