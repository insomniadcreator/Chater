package handlers

import (
	"log"
	"database/sql"
	"myapp/models"
	"net/http"
	"golang.org/x/crypto/bcrypt"
	"github.com/gorilla/sessions"
)


func RegisterHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
            username := r.FormValue("username")
            email := r.FormValue("email")
            password := r.FormValue("password")

            // Проверка уникальности email и username
            var count int
            err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1 OR username = $2", email, username).Scan(&count)
            if err != nil {
                log.Printf("Error checking uniqueness: %v", err)
                http.Error(w, "Database error", http.StatusInternalServerError)
                return
            }
            if count > 0 {
                http.Error(w, "Email or username already exists", http.StatusBadRequest)
                return
            }

            // Хеширование пароля
            hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
            if err != nil {
                http.Error(w, "Could not hash password", http.StatusInternalServerError)
                return
            }

            // Сохранение пользователя в базе данных
            _, err = db.Exec("INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)", username, email, string(hashedPassword))
            if err != nil {
                log.Printf("Error saving user: %v", err)
                http.Error(w, "Could not save user", http.StatusInternalServerError)
                return
            }

            // Перенаправление на страницу логина
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // Отображение страницы регистрации
        tmpl.ExecuteTemplate(w, "register.html", nil)
    }
}

func LoginHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
            username := r.FormValue("username")
            password := r.FormValue("password")

            // Поиск пользователя в базе данных
            var user models.User
            err := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.PasswordHash)
            if err != nil {
                if err == sql.ErrNoRows {
                    http.Error(w, "Invalid credentials", http.StatusUnauthorized)
                } else {
                    http.Error(w, "Database error", http.StatusInternalServerError)
                }
                return
            }

            // Проверка пароля
            err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
            if err != nil {
                http.Error(w, "Invalid credentials", http.StatusUnauthorized)
                return
            }

            // Создание сессии
            session, _ := store.Get(r, "session-name")
            session.Values["user_id"] = user.ID
            session.Save(r, w)

            // Перенаправление на главную страницу
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }

        // Отображение страницы логина
        tmpl.ExecuteTemplate(w, "login.html", nil)
    }
}

func LogoutHandler(store *sessions.CookieStore) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Получение сессии
        session, _ := store.Get(r, "session-name")

        // Удаление данных пользователя из сессии
        delete(session.Values, "user_id")
        session.Save(r, w)

        // Перенаправление на главную страницу
        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}
