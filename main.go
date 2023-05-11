package main

// database password HovaSihvR070GCJf
import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	supa "github.com/nedpals/supabase-go"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"

	"log"
	"net/http"
	"os"
)

// used for Redis
var ctx = context.Background()

type Lawyer struct {
	Lawyer_name  string `json:"lawyerName"`
	Lawyer_id    string `json:"lawyerId"`
	Lawyer_rate  int32  `json:"lawyerRate"`
	Lawyer_type  string `json:"lawyerType"`
	Lawyer_state string `json:"lawyerState"`
	Lawyer_about string `json:"lawyerAbout"`
}

type Message struct {
	Message_id   string `json:"messageId"`
	Sender_id    string `json:"senderId"`
	Recipient_id string `json:"recipientId"`
	Content      string `json:"content"`
	created_at   string `json:"timestamp"`
}
type Client struct {
	Client_id     string `json:"clientId"`
	Client_name   string `json:"clientName"`
	Client_budget int32  `json:"clientBudget"`
	Client_email  string `json:"clientEmail"`
	Client_state  string `json:"clientState"`

	client_dashboard clientDashboard `json:"clientDashboard"`
	Password         string          `json:"password"`
}

type Authenticate struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	UserType string `json:"userType"`
}

type Session struct {
	Sessionid string `json:"sessionId"`
	Userid    string `json:"userId"`
}
type Search struct {
	State      string `json:"state"`
	PageNumber string `json:"pageNumber"`
}

type clientDashboard struct {
}

var databaseURL = "postgres://postgres:38sUitIKWGHMNEQQ@db.dentosxmnrtcqboongyz.supabase.co:6543/postgres"

func main() {

	// Redis credentials

	server := echo.New()

	server.Use(middleware.Logger())
	server.Use(middleware.Recover())
	server.POST("/getLawyer", getLawyer)
	server.GET("/lawyers", getLawyers)
	server.GET("/clients", getClients)
	// find one lawyer
	server.GET("/lawyers/:id", getLawyer)
	// Find one client
	server.GET("/clients/:id", getClient)
	server.GET("/getChats", getChats)
	// authenticate a user
	server.POST("login", authenticateUser)
	server.POST("/createuser", createUser)
	server.POST("/logout", logout)
	server.POST("/authenticateUser", authenticateUser)
	server.POST("/getDashboard", getDashboard)
	server.POST("/getLawyersByState", getLawyersByState)
	server.POST("/getChats", getChats)
	server.POST("/getChat/:id", getChatThread)

	server.Logger.Fatal(server.Start(":1323"))
	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
func authenticateUser(server echo.Context) error {
	var authenticate Authenticate

	err := server.Bind(&authenticate)
	if err != nil {
		return err
	}
	fmt.Println(authenticate)

	opt, err := redis.ParseURL("redis://default:jZVE2nugSKxJGTRvJjsZvrlBR1ACKEaL@redis-17179.c44.us-east-1-2.ec2.cloud.redislabs.com:17179")
	if err != nil {
		panic(err)
	}
	rbd := redis.NewClient(opt)

	supabaseUrl := "https://dentosxmnrtcqboongyz.supabase.co"
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImRlbnRvc3htbnJ0Y3Fib29uZ3l6Iiwicm9sZSI6ImFub24iLCJpYXQiOjE2NzkxNTU2MTksImV4cCI6MTk5NDczMTYxOX0.IFSy5kt4bYj5i7gtWtA2BMVH2S9n6lcN05ggrxPmiKI"
	supabase := supa.CreateClient(supabaseUrl, supabaseKey)

	ctx := context.Background()
	user, err := supabase.Auth.SignIn(ctx, supa.UserCredentials{
		Email:    authenticate.Email,
		Password: authenticate.Password,
	})
	if err != nil {
		panic(err)
	}
	var session Session

	session.Sessionid = user.RefreshToken
	session.Userid = user.User.ID
	rbderr := rbd.Set(ctx, user.RefreshToken, user.User.ID, time.Hour*24).Err()
	if rbderr != nil {
		panic(err)
	}

	fmt.Println(user.User.Email)

	log.Println(session)
	sessionMarshalled, _ := json.Marshal(session)
	fmt.Println(sessionMarshalled)
	// TODO: Make sure session is actully being sent to nodejs
	return server.JSON(http.StatusOK, user.RefreshToken)

}

func createUser(server echo.Context) error {

	var client Client

	err := server.Bind(&client)
	if err != nil {
		return err
	}

	supabaseUrl := "https://dentosxmnrtcqboongyz.supabase.co"
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImRlbnRvc3htbnJ0Y3Fib29uZ3l6Iiwicm9sZSI6ImFub24iLCJpYXQiOjE2NzkxNTU2MTksImV4cCI6MTk5NDczMTYxOX0.IFSy5kt4bYj5i7gtWtA2BMVH2S9n6lcN05ggrxPmiKI"
	supabase := supa.CreateClient(supabaseUrl, supabaseKey)

	ctx := context.Background()
	user, err := supabase.Auth.SignUp(ctx, supa.UserCredentials{
		Email:    client.Client_email,
		Password: client.Password,
	})
	if err != nil {
		panic(err)
		return server.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}

	dbpool, err := pgxpool.New(context.Background(), databaseURL)
	defer dbpool.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	// TODO: Query for Inserting new client in database
	// Start a new database transaction
	tx, err := dbpool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Prepare the insert statement
	query := `
        INSERT INTO client (client_name, column2, column3)
        VALUES ($1, $2, $3)
    `
	_, err = tx.Exec(context.Background(), "insert_query", query)
	if err != nil {
		return err
	}

	// Commit the transaction
	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(user)
	return server.JSON(http.StatusOK, map[string]string{
		"message": fmt.Sprintf("User %s has created"),
	})
}

func logout(server echo.Context) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis-17179.c44.us-east-1-2.ec2.cloud.redislabs.com",
		Password: "jZVE2nugSKxJGTRvJjsZvrlBR1ACKEaL", // no password set
		DB:       0,                                  // use default DB
	})
	sessionID := server.FormValue("sessionID")

	// Delete the session from Redis database
	err := rdb.Del(server.Request().Context(), sessionID).Err()
	if err != nil {
		return err
	}

	// return a success response
	return server.JSON(http.StatusOK, map[string]string{
		"message": fmt.Sprintf("User %s has been logged out"),
	})
}

// Get a singular lawyer
func getLawyer(server echo.Context) error {
	fmt.Println("Running...")
	var lawyer Lawyer

	bindError := server.Bind(&lawyer)
	if bindError != nil {
		fmt.Fprintf(os.Stderr, "bind error")
		return bindError
	}

	//var databaseURL = "postgres://postgres:38sUitIKWGHMNEQQ@db.dentosxmnrtcqboongyz.supabase.co:6543/postgres"
	dbpool, err := pgxpool.New(context.Background(), databaseURL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	err = dbpool.QueryRow(context.Background(), "select * from lawyer where lawyer_id=$1", lawyer.Lawyer_id).Scan(&lawyer.Lawyer_id, &lawyer.Lawyer_name, &lawyer.Lawyer_rate, &lawyer.Lawyer_type, &lawyer.Lawyer_state, &lawyer.Lawyer_about)

	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	return server.JSON(http.StatusOK, lawyer)

}

func getDashboard(server echo.Context) error {

	return server.JSON(http.StatusOK, "hello")
}

func getLawyers(server echo.Context) error {
	fmt.Println("Running...")
	//var databaseURL = "postgres://postgres:38sUitIKWGHMNEQQ@db.dentosxmnrtcqboongyz.supabase.co:6543/postgres"
	dbpool, err := pgxpool.New(context.Background(), databaseURL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()
	rows, err := dbpool.Query(context.Background(), "select * FROM lawyer")
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	var lawyerList []Lawyer
	fmt.Println(rows)
	for rows.Next() {
		// TODO: Type map here
		var lawyer Lawyer

		err = rows.Scan(&lawyer.Lawyer_id, &lawyer.Lawyer_name, &lawyer.Lawyer_rate, &lawyer.Lawyer_type, &lawyer.Lawyer_state)
		if err != nil {
			log.Fatal("error while iterating dataset")
		}

		lawyerList = append(lawyerList, lawyer)

	}
	for _, la := range lawyerList {
		fmt.Printf("%s, %s, %d, %s", la.Lawyer_id, la.Lawyer_name, la.Lawyer_rate, la.Lawyer_type)
	}
	j, _ := json.Marshal(lawyerList)
	log.Println(string(j))
	return server.JSON(http.StatusOK, lawyerList)

}

// Search for lawyers based on state
func getLawyersByState(server echo.Context) error {
	fmt.Println("Running...")
	//var databaseURL = "postgres://postgres:38sUitIKWGHMNEQQ@db.dentosxmnrtcqboongyz.supabase.co:6543/postgres"
	dbpool, err := pgxpool.New(context.Background(), databaseURL)
	var search Search
	err = server.Bind(&search)
	if err != nil {
		return err
	}

	fmt.Println(search.State)

	if err != nil {
		// handle error
		fmt.Println("Error converting string to integer:", err)
		return nil
	}
	newInt, _ := strconv.ParseInt(search.PageNumber, 0, 64)

	queryOffset := (newInt - 1) * 10
	fmt.Println(queryOffset)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()
	rows, err := dbpool.Query(context.Background(), "select * FROM lawyer WHERE lawyer.lawyer_state = $1 ORDER BY lawyer_name desc LIMIT 10 OFFSET $2", search.State, queryOffset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	var lawyerList []Lawyer
	for rows.Next() {
		// TODO: Type map here
		var lawyer Lawyer

		err = rows.Scan(&lawyer.Lawyer_id, &lawyer.Lawyer_name, &lawyer.Lawyer_rate, &lawyer.Lawyer_type, &lawyer.Lawyer_state, &lawyer.Lawyer_about)
		if err != nil {
			log.Fatal("error while iterating dataset")
		}

		lawyerList = append(lawyerList, lawyer)

	}
	for _, la := range lawyerList {
		fmt.Printf("%s, %s, %d, %s", la.Lawyer_id, la.Lawyer_name, la.Lawyer_rate, la.Lawyer_type)
	}
	j, _ := json.Marshal(lawyerList)
	log.Println(string(j))
	return server.JSON(http.StatusOK, lawyerList)

}

func getClients(server echo.Context) error {
	fmt.Println("Running...")
	//var databaseURL = "postgres://postgres:HovaSihvR070GCJf@db.dentosxmnrtcqboongyz.supabase.co:6543/postgres"
	dbpool, err := pgxpool.New(context.Background(), databaseURL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()
	rows, err := dbpool.Query(context.Background(), "select * FROM client")
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	var clientList []Client
	fmt.Println(rows)
	for rows.Next() {
		// TODO: Type map here
		var client Client

		err = rows.Scan(&client.Client_id, &client.Client_name, &client.Client_budget, &client.Client_email)
		if err != nil {
			log.Fatal("error while iterating dataset")
		}

		clientList = append(clientList, client)

	}
	for _, cl := range clientList {
		fmt.Printf("%s, %s, %d, %s", cl.Client_id, cl.Client_name, cl.Client_budget, cl.Client_email)
	}
	return server.String(http.StatusOK, clientList[0].Client_name)

}
func getClient(server echo.Context) error {
	fmt.Println("Running...")
	var client Client

	err := server.Bind(&client)
	if err != nil {
		return err
	}

	var databaseURL = "postgresql://postgres:HovaSihvR070GCJf@db.dentosxmnrtcqboongyz.supabase.co:5432/postgres"
	dbpool, err := pgxpool.New(context.Background(), databaseURL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	err = dbpool.QueryRow(context.Background(), "select * from client where client_id=$1", client.Client_id).Scan(&client.Client_id, &client.Client_name, &client.Client_email, &client.Client_budget, &client.Client_state)

	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	return server.JSON(http.StatusOK, client)
}

func getChats(server echo.Context) error {

	fmt.Println("Running...")
	type User struct {
		userid string `db:"lawyer_id" db:"client_id"`
	}
	var user User

	err := server.Bind(&user)
	fmt.Println(user)
	//var databaseURL = "postgres://postgres:HovaSihvR070GCJf@db.dentosxmnrtcqboongyz.supabase.co:6543/postgres"
	dbpool, err := pgxpool.New(context.Background(), databaseURL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)

	}
	defer dbpool.Close()
	rows, err := dbpool.Query(context.Background(), "SELECT u.user_id, u.username, m.created_at, m.content "+
		"FROM ("+
		"SELECT client_id as user_id, client_username as username FROM client "+
		"UNION "+
		"SELECT lawyer_id as user_id, lawyer_username as username FROM lawyer "+
		") u "+
		"JOIN messages m ON (u.user_id = m.sender_id OR u.user_id = m.recipient_id) "+
		"WHERE (m.sender_id = $1 OR m.recipient_id = $1) AND u.user_id != $1 GROUP BY u.user_id, u.username, m.created_at, m.content HAVING m.created_at = MAX(m.created_at) ORDER BY m.created_at DESC;", user.userid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	var messageList []Message
	fmt.Println(rows)
	for rows.Next() {
		// Type map here
		var message Message

		err = rows.Scan(&message.Message_id, &message.Sender_id, &message.Content, &message.Recipient_id, &message.created_at)

		if err != nil {
			log.Fatal("error while iterating dataset")
		}

		messageList = append(messageList, message)

	}
	for _, ms := range messageList {
		fmt.Printf("%s, %s, %s, %s", ms.Recipient_id, ms.Sender_id, ms.Content, ms.created_at)
	}

	return server.JSON(http.StatusOK, messageList)

}

func getChatThread(server echo.Context) error {

	fmt.Println("Running...")
	type User struct {
		userid string `db:"lawyer_id" db:"client_id"`
	}
	var user User

	err := server.Bind(&user)
	fmt.Println(user)
	//var databaseURL = "postgres://postgres:HovaSihvR070GCJf@db.dentosxmnrtcqboongyz.supabase.co:6543/postgres"
	dbpool, err := pgxpool.New(context.Background(), databaseURL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)

	}
	defer dbpool.Close()
	rows, err := dbpool.Query(context.Background(), "SELECT * FROM messages WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?) ORDER BY created_at DESC, sender_id, recipient_id, recipient_ID, sender_id", user.userid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	var messageList []Message
	fmt.Println(rows)
	for rows.Next() {
		// Type map here
		var message Message

		err = rows.Scan(&message.Message_id, &message.Sender_id, &message.Content, &message.Recipient_id, &message.created_at)

		if err != nil {
			log.Fatal("error while iterating dataset")
		}

		messageList = append(messageList, message)

	}
	for _, ms := range messageList {
		fmt.Printf("%s, %s, %s, %s", ms.Recipient_id, ms.Sender_id, ms.Content, ms.created_at)
	}

	return server.JSON(http.StatusOK, messageList)

}
