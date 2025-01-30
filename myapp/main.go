package main

import (
	"log"
	"myapp/database"
	"myapp/handlers"
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("your-secret-key"))

func main() {
    // Подключение к базе данных
    connStr := "user=postgres dbname=chat-app sslmode=disable password=postgres host=localhost port=5432"
    database.InitDB(connStr)

    // Настройка маршрутов
    http.HandleFunc("/", handlers.HomeHandler(database.DB, store))
	http.HandleFunc("/feedback", handlers.FeedbackHandler(database.DB, store))
    http.HandleFunc("/edit-message/", handlers.EditMessageHandler(database.DB, store))
http.HandleFunc("/update-message/", handlers.UpdateMessageHandler(database.DB, store))
	http.HandleFunc("/submit-message", handlers.SubmitMessageHandler(database.DB, store))
    http.HandleFunc("/register", handlers.RegisterHandler(database.DB, store))
    http.HandleFunc("/login", handlers.LoginHandler(database.DB, store))
    http.HandleFunc("/logout", handlers.LogoutHandler(store))
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    http.HandleFunc("/delete-message", handlers.DeleteMessageHandler(database.DB, store))

    // Запуск сервера
    log.Println("Server is up and running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}


