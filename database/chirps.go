package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func (db *DB) CreateUser(email string, password string) (UserResponse, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	user := User{
		ID:       db.nextUserID,
		Email:    email,
		Password: string(hashedPassword),
	}

	userResponse := UserResponse{
		ID:    db.nextUserID,
		Email: email,
	}

	for _, userData := range db.users {
		if user.Email == userData.Email {
			return UserResponse{}, errors.New("email already in use")
		}
	}

	db.users[user.ID] = user
	db.usersResponse[user.ID] = userResponse

	db.nextUserID++

	if err := db.writeDB(); err != nil {
		return userResponse, err
	}

	return userResponse, nil
}

func (db *DB) GetUsers() ([]User, []UserResponse, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	users := make([]User, 0, len(db.users))
	for _, user := range db.users {
		users = append(users, user)
	}

	usersResponse := make([]UserResponse, 0, len(db.usersResponse))
	for _, userResponse := range db.usersResponse {
		usersResponse = append(usersResponse, userResponse)
	}

	return users, usersResponse, nil
}
