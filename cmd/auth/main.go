package main

import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/golang-jwt/jwt/v5"
    "github.com/nbd-wtf/go-nostr"
    "github.com/redis/go-redis/v9"
)

var (
    jwtSecret   = []byte("your-secret-key") // From env
    redisClient = redis.NewClient(&redis.Options{Addr: "redis:6379"})
)

type LoginRequest struct {
    Event nostr.Event `json:"event"`
}

type LoginResponse struct {
    Token string `json:"token"`
}

func login(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    if !req.Event.Verify() {
        http.Error(w, "Invalid Nostr signature", http.StatusUnauthorized)
        return
    }

    if req.Event.Kind != 1 || req.Event.Content != "login" {
        http.Error(w, "Invalid event", http.StatusBadRequest)
        return
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "pubkey": req.Event.PubKey,
        "exp":    time.Now().Add(24 * time.Hour).Unix(),
    })
    tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        return
    }

    err = redisClient.Set(context.Background(), "session:"+req.Event.PubKey, tokenString, 24*time.Hour).Err()
    if err != nil {
        http.Error(w, "Failed to store session", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(LoginResponse{Token: tokenString})
}

func main() {
    router := mux.NewRouter()
    router.HandleFunc("/auth/login", login).Methods("POST")
    log.Fatal(http.ListenAndServe(":8080", router))
}
