package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

type UserResponse struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

func (db *DB) CreateUser(email string, password string) (UserResponse, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	user := User{
		ID:          db.nextUserID,
		Email:       email,
		Password:    string(hashedPassword),
		IsChirpyRed: false,
	}

	userResponse := UserResponse{
		ID:          db.nextUserID,
		Email:       email,
		IsChirpyRed: false,
	}

	for _, userData := range db.users {
		if user.Email == userData.Email {
			return UserResponse{}, errors.New("email already in use")
		}
	}

	db.users[user.ID] = user

	db.nextUserID++

	if err := db.writeDB(); err != nil {
		return UserResponse{}, err
	}

	return userResponse, nil
}

func (db *DB) GetUsers() ([]User, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	users := make([]User, 0, len(db.users))
	for _, user := range db.users {
		users = append(users, user)
	}

	return users, nil
}

func (db *DB) UpdateUser(userID int, email string, password string, ischirpyred bool, usingWebhook bool) (User, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	if !usingWebhook {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}
		password = string(hashedPassword)
	}

	user := User{
		ID:          userID,
		Email:       email,
		Password:    password,
		IsChirpyRed: ischirpyred,
	}

	db.users[userID] = user

	return user, nil
}
