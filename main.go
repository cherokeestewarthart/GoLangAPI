package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	_ "github.com/go-sql-driver/mysql"
)

type Application struct {
	DB *sql.DB
}

func writeError(w http.ResponseWriter, err error) {
	fmt.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(nil)
}

func createDatabase() *sql.DB {
	db, err := sql.Open("mysql", "root:Cast7371@/Grupo?parseTime=true")
	if err != nil {
		panic(err.Error())
	}

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	return db
}

// Credentials Info
type Credentials struct {
	Cid     string `json:clientid`
	Csecret string `json:csecret`
	// ProjectID string `json:"grupoapi-186217"`
	// Csecret string `json:"BgsuRaKdCLHI51ySU6p-zEfp"`
	// Csecret string `json:"BgsuRaKdCLHI51ySU6p-zEfp"`
}

//"auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://accounts.google.com/o/oauth2/token",
//"auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs",
//"client_secret":"BgsuRaKdCLHI51ySU6p-zEfp","redirect_uris":["http://127.0.0.1:8080"]}}
// GmailUser Info
type GmailUser struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Gender        string `json:"gender"`
}

var cred Credentials
var state string
var store = sessions.NewCookieStore([]byte("secret"))

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

var (
	conf *oauth2.Config
)

func init() {
	cred := Credentials{
		Cid:     "411040242308-qkl164l03pfeukholtaur1nmdvvcjll5.apps.googleusercontent.com",
		Csecret: "BgsuRaKdCLHI51ySU6p-zEfp",
	}

	conf = &oauth2.Config{
		ClientID:     cred.Cid,
		ClientSecret: cred.Csecret,
		RedirectURL:  "http://127.0.0.1:8080",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email", // You have to select your own scope from here -> https://developers.google.com/identity/protocols/googlescopes#google_sign-in
		},
		Endpoint: google.Endpoint,
	}

}
func getLoginURL(state string) string {
	// State can be some kind of random generated hash string.
	// See relevant RFC: http://tools.ietf.org/html/rfc6749#section-10.12
	return conf.AuthCodeURL(state)
}
func (app *Application) AuthHandler(c *gin.Context) {
	// Handle the exchange code to initiate a transport.
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	queryState := c.Request.URL.Query().Get("state")
	if retrievedState != queryState {
		log.Printf("Invalid session state: retrieved: %s; Param: %s", retrievedState, queryState)
		c.HTML(http.StatusUnauthorized, "error.tmpl", gin.H{"message": "Invalid session state."})
		return
	}
	code := c.Request.URL.Query().Get("code")
	tok, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Login failed. Please try again."})
		return
	}

	client := conf.Client(oauth2.NoContext, tok)
	userinfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	defer userinfo.Body.Close()
	data, _ := ioutil.ReadAll(userinfo.Body)
	u := GmailUser{}
	if err = json.Unmarshal(data, &u); err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error marshalling response. Please try agian."})
		return
	}
	session.Set("user-id", u.Email)
	err = session.Save()
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error while saving session. Please try again."})
		return
	}
	seen := false
	app.SaveUser(&User{
		Name:  u.Name,
		Email: u.Email,
	})
	// if _, mongoErr := db.LoadUser(u.Email); mongoErr == nil {
	// 	seen = true
	// } else {
	// 	err = db.SaveUser(&u)
	// 	if err != nil {
	// 		log.Println(err)
	// 		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error while saving user. Please try again."})
	// 		return
	// 	}
	// }
	c.HTML(http.StatusOK, "battle.tmpl", gin.H{"email": u.Email, "seen": seen})
}

func loginHandler(c *gin.Context) {
	state = randToken()
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()
	c.Writer.Write([]byte("<html><title>Golang Google</title> <body> <a href='" + getLoginURL(state) + "'><button>Login with Google!</button> </a> </body></html>"))
}
func main() {
	db := createDatabase()
	defer db.Close()
	var myTime time.Time
	rows, err := db.Query("SELECT current_timestamp()")

	if rows.Next() {
		if err = rows.Scan(&myTime); err != nil {
			panic(err)
		}
	}

	fmt.Println(myTime)
	app := &Application{
		DB: db,
	}
	_ = app
	router := gin.Default()

	router.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.String(http.StatusOK, "Hello %s", name)
	})

	router.Use(sessions.Sessions("goquestsession", store))
	router.Static("/css", "./static/css")
	router.Static("/img", "./static/img")
	router.LoadHTMLGlob("templates/*")

	router.GET("/login", loginHandler)
	router.GET("/auth", app.AuthHandler)
	router.Any("/user", app.UserHandler)
	router.Any("/groups", app.GroupHandler)
	router.Any("/createaccount", app.AccountHandler)
	fmt.Println("Listening...")
	log.Fatal(router.Run("127.0.0.1:8080"))
}

// execute Application.UserHandler
func (app *Application) UserHandler(c *gin.Context) {
	var data []byte
	r := c.Request
	w := c.Writer
	switch r.Method {
	case http.MethodGet:
		users, err := app.GetAllUsers()
		if err != nil {
			writeError(w, err)
			return
		}
		data, err = json.Marshal(users)
		if err != nil {
			writeError(w, err)
			return
		}
		break
	case http.MethodPost:
		newUser := &User{}
		rawBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeError(w, err)
			return
		}
		err = json.Unmarshal(rawBody, newUser)
		if err != nil {
			writeError(w, err)
			return
		}
		user, err := app.SaveUser(newUser)
		if err != nil {
			writeError(w, err)
			return
		}
		data, err = json.Marshal(user)
		if err != nil {
			writeError(w, err)
			return
		}
		break
	}

	w.Write(data)
}

// exported method  Application.GroupHandler
func (app *Application) GroupHandler(c *gin.Context) {
	var data []byte
	w := c.Writer
	r := c.Request

	switch r.Method {
	case http.MethodGet:
		Groups, err := app.GetAllGroups()
		if err != nil {
			writeError(w, err)
			return
		}
		data, err = json.Marshal(Groups)
		if err != nil {
			writeError(w, err)
			return
		}
		break
	case http.MethodPost:
		newGroup := &Groups{}
		rawBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeError(w, err)
			return
		}
		err = json.Unmarshal(rawBody, newGroup)
		if err != nil {
			writeError(w, err)
			return
		}
		Groups, err := app.SaveGroup(newGroup)
		if err != nil {
			writeError(w, err)
			return
		}
		data, err = json.Marshal(Groups)
		if err != nil {
			writeError(w, err)
			return
		}
		break
	}

	w.Write(data)
}

// exported method  AccountHandler
func (app *Application) AccountHandler(c *gin.Context) {
	var data []byte
	w := c.Writer
	r := c.Request

	switch r.Method {
	case http.MethodGet:
		CreateAccount, err := app.GetAllAccount()
		if err != nil {
			writeError(w, err)
			return
		}
		data, err = json.Marshal(CreateAccount)
		if err != nil {
			writeError(w, err)
			return
		}
		break
	case http.MethodPost:
		newAccount := &CreateAccount{}
		rawBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeError(w, err)
			return
		}
		err = json.Unmarshal(rawBody, newAccount)
		if err != nil {
			writeError(w, err)
			return
		}
		CreateAccount, err := app.SaveAccount(newAccount)
		if err != nil {
			writeError(w, err)
			return
		}
		data, err = json.Marshal(CreateAccount)
		if err != nil {
			writeError(w, err)
			return
		}
		break
	}

	w.Write(data)
}

//client secret : BgsuRaKdCLHI51ySU6p-zEfp
//client ID : 411040242308-qkl164l03pfeukholtaur1nmdvvcjll5.apps.googleusercontent.com
