package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

var domain string
var port string

type YAMLFile struct {
	Config Config `yaml:"config"`
}

type Config struct {
	MYSQL_DB      string `yaml:"MYSQL_DB"`
	MYSQL_USER    string `yaml:"MYSQL_USER"`
	MYSQL_PASS    string `yaml:"MYSQL_PASS"`
	MYSQL_HOST    string `yaml:"MYSQL_HOST"`
	DOMAIN        string `yaml:"DOMAIN"`
	PORT          string `yaml:"PORT"`
	DISCORD_TOKEN string `yaml:"DISCORD_TOKEN"`
}

func ReadConfig() (*Config, error) {
	config := &YAMLFile{}
	cfgFile, err := os.ReadFile("./config.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(cfgFile, config)
	return &config.Config, err
}

type user struct {
	Id         int    `json:"id"`
	CreatedAt  string `json:"created_at"`
	DiscordId  string `json:"discord_id"`
	FriendCode string `json:"friend_code"`
}

func ReturnSingleUser(w http.ResponseWriter, r *http.Request) {
	var u user
	vars := mux.Vars(r)
	id := vars["id"]

	query := `SELECT id, created_at, discord_id, friend_code FROM users where discord_id = ?`
	err := db.QueryRow(query, id).Scan(&u.Id, &u.CreatedAt, &u.DiscordId, &u.FriendCode)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad Request")
		return
	}

	json.NewEncoder(w).Encode(u)
}

func GetMultipleUsers(w http.ResponseWriter, r *http.Request) {
	// Read body
	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Unmarshal
	var getUsers []user
	err = json.Unmarshal(b, &getUsers)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	var returnUsers []user

	for _, us := range getUsers {
		var u user
		query := `SELECT id, created_at, discord_id, friend_code FROM users where discord_id = ?`
		err := db.QueryRow(query, us.DiscordId).Scan(&u.Id, &u.CreatedAt, &u.DiscordId, &u.FriendCode)
		if err != nil {
			log.Print(err)
		}
		returnUsers = append(returnUsers, u)
	}

	json.NewEncoder(w).Encode(returnUsers)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.URL.Host != "localhost:8383" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Unauthorized")
		return
	}
	// Read body
	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Unmarshal
	var msg user
	err = json.Unmarshal(b, &msg)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	_, err = db.Exec(`INSERT INTO users(discord_id, friend_code) VALUES (?, ?)`, msg.DiscordId, msg.FriendCode)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request")
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "OK")
}

func CreateToken(w http.ResponseWriter, r *http.Request) {
	if r.Host != "localhost:8383" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Unauthorized")
		return
	}
	token, err := GenerateRandomString(8)
	if err != nil {
		return
	}

	vars := mux.Vars(r)
	discord := vars["discord"]
	if len(discord) != 18 {
		fmt.Fprintf(w, "Bad request")
		return
	}

	_, err = db.Exec(`INSERT INTO tokens(discord_id, token) VALUES (?, ?)`, discord, token)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request")
		return
	}

	json.NewEncoder(w).Encode(token)
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

var bearer string

var db *sql.DB

func main() {
	var err error
	config, err := ReadConfig()
	if err != nil {
		log.Fatal(err)
	}
	domain = config.DOMAIN
	port = config.PORT

	db, err = sql.Open("mysql", config.MYSQL_USER+":"+config.MYSQL_PASS+"@("+config.MYSQL_HOST+")/"+config.MYSQL_DB+"?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Database connection established")

	bearer = "Bot " + config.DISCORD_TOKEN

	router := mux.NewRouter()

	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/discover", discoverHandler)
	router.HandleFunc("/openidcallback", callbackHandler)
	router.HandleFunc("/{token}", indexHandler)
	// router.HandleFunc("/api/v1/user/{id}", ReturnSingleUser).Methods("GET")
	router.HandleFunc("/api/v1/users", GetMultipleUsers)
	//router.HandleFunc("/api/v1/user", CreateUser).Methods("POST")
	router.HandleFunc("/api/v1/token/{discord}", CreateToken)
	http.ListenAndServe(port, router)
}
