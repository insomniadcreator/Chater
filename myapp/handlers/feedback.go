package handlers

import (
    "database/sql"
    "net/http"
    "github.com/gorilla/sessions"
    "myapp/models"
)

func FeedbackHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Получаем сессию пользователя
        session, err := store.Get(r, "session-name")
        if err != nil {
            http.Error(w, "Session error", http.StatusInternalServerError)
            return
        }

        // Получаем ID пользователя из сессии
        userID, ok := session.Values["user_id"].(int)
        if !ok {
            userID = 0 // Если пользователь не авторизован, используем 0
        }

        // Получаем имя пользователя из базы данных
        var username string
        if userID != 0 {
            err := db.QueryRow("SELECT username FROM users WHERE id = $1", userID).Scan(&username)
            if err != nil {
                http.Error(w, "Could not fetch username", http.StatusInternalServerError)
                return
            }
        }

        // Получаем параметр поиска из запроса
        searchQuery := r.URL.Query().Get("search")

        // Получаем сообщения из базы данных
        rows, err := db.Query(`
            SELECT m.id, m.user_id, m.message, m.created_at, u.username
            FROM messages m
            JOIN users u ON m.user_id = u.id
            ORDER BY m.created_at DESC
        `)
        if err != nil {
            http.Error(w, "Could not fetch messages", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        // Собираем сообщения в список
        var messages []models.Message
        for rows.Next() {
            var msg models.Message
            err := rows.Scan(&msg.ID, &msg.UserID, &msg.Message, &msg.CreatedAt, &msg.Username)
            if err != nil {
                http.Error(w, "Could not read message", http.StatusInternalServerError)
                return
            }
            messages = append(messages, msg)
        }

        // Получаем список всех пользователей с учетом поиска
        var usersRows *sql.Rows
        if searchQuery != "" {
            // Поиск по юзернейму
            usersRows, err = db.Query("SELECT username, email FROM users WHERE username ILIKE $1", "%"+searchQuery+"%")
        } else {
            // Все пользователи
            usersRows, err = db.Query("SELECT username, email FROM users")
        }
        if err != nil {
            http.Error(w, "Could not fetch users", http.StatusInternalServerError)
            return
        }
        defer usersRows.Close()

        // Собираем пользователей в список
        var users []models.User
        for usersRows.Next() {
            var user models.User
            err := usersRows.Scan(&user.Username, &user.Email)
            if err != nil {
                http.Error(w, "Could not read user", http.StatusInternalServerError)
                return
            }
            users = append(users, user)
        }

        // Отображаем страницу с сообщениями и пользователями
        tmpl.ExecuteTemplate(w, "feedback.html", map[string]interface{}{
            "Messages":    messages,
            "UserID":      userID,
            "Username":    username,
            "Users":       users,
            "SearchQuery": searchQuery, // Передаем поисковый запрос в шаблон
        })
    }
}

func SubmitMessageHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
            // Получаем ID пользователя из сессии
            session, _ := store.Get(r, "session-name")
            userID, ok := session.Values["user_id"].(int)
            if !ok {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            // Получаем текст сообщения из формы
            message := r.FormValue("message")

            // Сохраняем сообщение в базе данных
            _, err := db.Exec("INSERT INTO messages (user_id, message) VALUES ($1, $2)", userID, message)
            if err != nil {
                http.Error(w, "Could not save message", http.StatusInternalServerError)
                return
            }

            // Перенаправляем пользователя на страницу чата
            http.Redirect(w, r, "/feedback", http.StatusSeeOther)
            return
        }

        // Если метод не POST, возвращаем ошибку
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
    }
}

func DeleteMessageHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
            // Получаем ID пользователя из сессии
            session, _ := store.Get(r, "session-name")
            userID, ok := session.Values["user_id"].(int)
            if !ok {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            // Получаем ID сообщения из формы
            messageID := r.FormValue("message_id")

            // Проверяем, что сообщение принадлежит пользователю
            var msgUserID int
            err := db.QueryRow("SELECT user_id FROM messages WHERE id = $1", messageID).Scan(&msgUserID)
            if err != nil {
                http.Error(w, "Message not found", http.StatusNotFound)
                return
            }

            if msgUserID != userID {
                http.Error(w, "You can only delete your own messages", http.StatusForbidden)
                return
            }

            // Удаляем сообщение из базы данных
            _, err = db.Exec("DELETE FROM messages WHERE id = $1", messageID)
            if err != nil {
                http.Error(w, "Could not delete message", http.StatusInternalServerError)
                return
            }

            // Перенаправляем пользователя на страницу чата
            http.Redirect(w, r, "/feedback", http.StatusSeeOther)
            return
        }

        // Если метод не POST, возвращаем ошибку
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
    }
}

func EditMessageHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Получаем ID сообщения из URL
        messageID := r.URL.Path[len("/edit-message/"):]

        // Получаем сессию пользователя
        session, _ := store.Get(r, "session-name")
        userID, ok := session.Values["user_id"].(int)
        if !ok {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Проверяем, что сообщение принадлежит пользователю
        var msgUserID int
        var messageText string
        err := db.QueryRow("SELECT user_id, message FROM messages WHERE id = $1", messageID).Scan(&msgUserID, &messageText)
        if err != nil {
            http.Error(w, "Message not found", http.StatusNotFound)
            return
        }

        if msgUserID != userID {
            http.Error(w, "You can only edit your own messages", http.StatusForbidden)
            return
        }

        // Отображаем страницу редактирования
        tmpl.ExecuteTemplate(w, "edit-message.html", map[string]interface{}{
            "MessageID":   messageID,
            "MessageText": messageText,
        })
    }
}

func UpdateMessageHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
            // Получаем ID сообщения из URL
            messageID := r.URL.Path[len("/update-message/"):]

            // Получаем сессию пользователя
            session, _ := store.Get(r, "session-name")
            userID, ok := session.Values["user_id"].(int)
            if !ok {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            // Проверяем, что сообщение принадлежит пользователю
            var msgUserID int
            err := db.QueryRow("SELECT user_id FROM messages WHERE id = $1", messageID).Scan(&msgUserID)
            if err != nil {
                http.Error(w, "Message not found", http.StatusNotFound)
                return
            }

            if msgUserID != userID {
                http.Error(w, "You can only edit your own messages", http.StatusForbidden)
                return
            }

            // Получаем новый текст сообщения из формы
            newMessage := r.FormValue("message")

            // Обновляем сообщение в базе данных
            _, err = db.Exec("UPDATE messages SET message = $1 WHERE id = $2", newMessage, messageID)
            if err != nil {
                http.Error(w, "Could not update message", http.StatusInternalServerError)
                return
            }

            // Перенаправляем пользователя на страницу чата
            http.Redirect(w, r, "/feedback", http.StatusSeeOther)
            return
        }

        // Если метод не POST, возвращаем ошибку
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
    }
}