package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

type User struct {
	UserName string    `json:"userName"`
	UserId   uuid.UUID `json:"userId"`
}

type UserStore struct {
	Users map[string]User `json:"users"`
	mu    sync.RWMutex
	file  string
}

func NewUserStore(file string) (*UserStore, error) {
	store := &UserStore{
		Users: make(map[string]User),
		file:  file,
	}
	err := store.Load()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return store, nil
}

func (s *UserStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.Users)
}

func (s *UserStore) Save() error {

	data, err := json.MarshalIndent(s.Users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.file, data, 0644)
}

func (s *UserStore) AddUser(username string) (*User, error) {
	if username == "" {
		return nil, errors.New("user name can't be empty")
	}
	if _, exists := s.Users[username]; exists {
		return nil, errors.New("user already exists")
	}
	newUser := User{UserName: username, UserId: uuid.New()}
	s.Users[username] = newUser
	fmt.Println(s.Users)
	err := s.Save()
	return &newUser, err
}

func (s *UserStore) GetUser(username string) (User, bool) {
	user, exists := s.Users[username]
	return user, exists
}
func (s *UserStore) GetUserById(userId string) (User, bool) {
	for _, v := range s.Users {
		if v.UserId.String() == userId {
			return v, true
		}
	}
	return User{}, false
}
