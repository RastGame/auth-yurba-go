package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
)

// Configuration
const (
    publicKey   = "" // public key application
    secretKey   = "" // secret key application
    redirectURL = "http://localhost:3000/redirect" // redirect url application
    api      = "https://api.yurba.one/" // api url yurba  
)

type ShortUserModel struct { // -> https://docs.yurba.one/get_user#shortusermodel
    ID                 int    `json:"ID"`
    Name               string `json:"Name"`
    Surname            string `json:"Surname"`
    Link               string `json:"Link"`
    Avatar             int    `json:"Avatar"`
    Sub                int    `json:"Sub"`
    Verify             string `json:"Verify"`
    Ban                int    `json:"Ban"`
    Emoji              string `json:"Emoji"`
    CosmeticAvatar     int    `json:"CosmeticAvatar"`
    CommentsState      int    `json:"CommentsState"`
    RelationshipState  string `json:"RelationshipState"`
}

func main() {
    http.HandleFunc("/", loginHandler)
    http.HandleFunc("/redirect", redirectHandler)

    fmt.Println("Server started at :3000")
    log.Println("Starting server at port 3000")
    
    if err := http.ListenAndServe(":3000", nil); err != nil {
        log.Fatalf("Error starting server: %v", err)
        os.Exit(1)
    }
}

// Redirect the user to Yurba for authentication -> https://docs.yurba.one/oauth
func loginHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Login request received")
    url := fmt.Sprintf("https://yurba.one/login/?publicKey=%s&redirectUrl=%s", publicKey, redirectURL)
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
    log.Println("Redirecting to Yurba for authentication")
}

// Handle the redirect from Yurba
func redirectHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Redirect request received from Yurba")
    if r.URL.Query().Get("success") == "1" {
        token := r.URL.Query().Get("token")
        log.Printf("Authentication successful. Token received: %s\n", token)

        if token != "" {
            user, err := getUserFromToken(token)
            if err != nil {
                http.Error(w, "Error fetching user: "+err.Error(), http.StatusInternalServerError)
                log.Printf("Error fetching user: %v\n", err)
                return
            }
            // Here, you can handle the authenticated user (e.g., store in session, etc.)
            fmt.Fprintf(w, "Authenticated User: %+v\n", user)
            log.Printf("Authenticated User: %+v\n", user)
            return
        }
    }
    http.Error(w, "Authentication failed", http.StatusUnauthorized)
    log.Println("Authentication failed: Token not provided or invalid")
}

// Fetch user information using the provided token -> https://docs.yurba.one/app_get_user
func getUserFromToken(token string) (ShortUserModel, error) {
    log.Printf("Fetching user information for token: %s\n", token)
    
    req, err := http.NewRequest("GET", api+"apps/user/"+token, nil)
    if err != nil {
        log.Printf("Error creating request: %v\n", err)
        return ShortUserModel{}, err
    }
    req.Header.Set("Secret-Key", secretKey)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error sending request: %v\n", err)
        return ShortUserModel{}, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Printf("Failed to get user: %s\n", resp.Status)
        return ShortUserModel{}, fmt.Errorf("failed to get user: %s", resp.Status)
    }

    var user ShortUserModel
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        log.Printf("Error decoding response: %v\n", err)
        return ShortUserModel{}, err
    }
    
    log.Printf("User information fetched successfully: %+v\n", user)
    return user, nil
}
