package database

import (
	"errors"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

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

func (db *DB) UpdateUser(ID string, email string, password string) (User, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	userID, err := strconv.Atoi(ID)
	if err != nil {
		panic(err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	user := User{
		ID:       userID,
		Email:    email,
		Password: string(hashedPassword),
	}

	db.users[userID] = user

	return user, nil
}
