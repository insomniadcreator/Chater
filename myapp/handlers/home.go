package handlers

import (
    "log"
    "database/sql"
	"net/http"
    "github.com/gorilla/sessions"
)

func HomeHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Получение сессии
        session, _ := store.Get(r, "session-name")
        userID, ok := session.Values["user_id"].(int)

        if !ok {
            // Пользователь не авторизован
            tmpl.ExecuteTemplate(w, "home.html", map[string]interface{}{
                "IsAuthenticated": false,
            })
            return
        }

        // Пользователь авторизован
        var username string
        err := db.QueryRow("SELECT username FROM users WHERE id = $1", userID).Scan(&username)
        if err != nil {
            log.Printf("Error fetching username: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        tmpl.ExecuteTemplate(w, "home.html", map[string]interface{}{
            "IsAuthenticated": true,
            "Username":        username,
        })
    }
}