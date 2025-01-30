package models

import "time"

type Message struct {
    ID        int       `json:"id"`
    UserID    int       `json:"user_id"`
    Message   string    `json:"message"`
    CreatedAt time.Time `json:"created_at"`
    Username  string    `json:"username"` // Имя пользователя (для отображения)
}