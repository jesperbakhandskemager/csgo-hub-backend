package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emily33901/go-csfriendcode"
	"github.com/gorilla/mux"
	"github.com/yohcop/openid-go"
)

// Load the templates once
var templateDir = "./"
var indexTemplate = template.Must(template.ParseFiles(templateDir + "index.html"))

// NoOpDiscoveryCache implements the DiscoveryCache interface and doesn't cache anything.
// For a simple website, I'm not sure you need a cache.
type NoOpDiscoveryCache struct{}

// Put is a no op.
func (n *NoOpDiscoveryCache) Put(id string, info openid.DiscoveredInfo) {}

// Get always returns nil.
func (n *NoOpDiscoveryCache) Get(id string) openid.DiscoveredInfo {
	return nil
}

var nonceStore = openid.NewSimpleNonceStore()
var discoveryCache = &NoOpDiscoveryCache{}

type IndexStruct struct {
	DiscordName   string `json:"DiscordName"`
	DiscordAvatar string `json:"DiscordAvatar"`
}

type DiscordUser struct {
	Id               string `json:"id"`
	Username         string `json:"username"`
	Avatar           string `json:"avatar"`
	AvatarDecoration string `json:"avatar_decoration"`
	Discriminator    string `json:"discriminator"`
	PublicFlags      int    `json:"public_flags"`
	Banner           string `json:"banner"`
	BannerColor      string `json:"banner_color"`
	AccentColor      string `json:"accent_color"`
}

// indexHandler serves up the index template with the "Sign in through STEAM" button.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	var discordId string
	query := `SELECT discord_id, token FROM tokens where token = BINARY ?`
	err := db.QueryRow(query, token).Scan(&discordId, &token)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request")
		return
	}
	// 959336363172442152/1ce1214a9540ff02cedc0acd0ad37d1f.png
	req, _ := http.NewRequest("GET", "https://discord.com/api/v9/users/"+discordId, nil)
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERROR] -", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
	}
	var discord DiscordUser
	json.Unmarshal(body, &discord)
	var discordAvatarURL string
	if discord.Avatar == "" {
		discordAvatarURL = "https://csgohub.xyz/assets/empty-avatar.png"
	} else {
		discordAvatarURL = "https://cdn.discordapp.com/avatars/" + discord.Id + "/" + discord.Avatar + ".png?size=100"
	}
	tmpl := IndexStruct{DiscordName: discord.Username, DiscordAvatar: discordAvatarURL}
	log.Println(token)
	expiration := time.Now().Add(time.Hour)
	cookie := http.Cookie{Name: "token", Value: token, Expires: expiration}
	http.SetCookie(w, &cookie)
	indexTemplate.Execute(w, tmpl)
}

// discoverHandler calls the Steam openid API and redirects to steam for login.
func discoverHandler(w http.ResponseWriter, r *http.Request) {
	url, err := openid.RedirectURL(
		"http://steamcommunity.com/openid",
		"http://"+domain+"/openidcallback",
		"http://"+domain+"/")

	if err != nil {
		log.Printf("Error creating redirect URL: %q\n", err)
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

func ClearToken(token string) error {
	_, err := db.Exec(`DELETE FROM tokens where token = ?`, token)
	return err
}

// callbackHandler handles the response back from Steam. It verifies the callback and then renders
// the index template with the logged in user's id.
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	fullURL := "http://" + domain + r.URL.String()

	id, err := openid.Verify(fullURL, discoveryCache, nonceStore)
	if err != nil {
		log.Printf("Error verifying: %q\n", err)
	} else {
		log.Printf("NonceStore: %+v\n", nonceStore)
		data := make(map[string]string)
		println(id)
		steamId, err := strconv.Atoi(strings.Split(id, "/id/")[1])
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Bad request")
			return
		}

		friendCode := csfriendcode.Encode(uint64(steamId))
		cookie, _ := r.Cookie("token")
		token := cookie.Value
		var discordId string

		query := `SELECT discord_id FROM tokens where token = ?`
		err = db.QueryRow(query, token).Scan(&discordId)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Bad request")
			return
		}
		ClearToken(token)

		var checkId, checkFriend string
		query = `SELECT discord_id, friend_code FROM users where discord_id = ?`
		err = db.QueryRow(query, discordId).Scan(&checkId, &checkFriend)
		if err == nil {
			_, err = db.Exec(`UPDATE users SET friend_code = ? WHERE discord_id = ?`, friendCode, discordId)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Bad request")
			}
			log.Print(err)
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, "Account link updated")
			return
		}

		_, err = db.Exec(`INSERT INTO users(discord_id, friend_code) VALUES (?, ?)`, discordId, friendCode)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Bad request")
		}

		w.WriteHeader(http.StatusCreated)
		data["user"] = id
		indexTemplate.Execute(w, data)
	}
}
