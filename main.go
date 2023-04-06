package main

// database password HovaSihvR070GCJf
import (
	"context"
	"encoding/json"

	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"log"
	"net/http"
	"os"
)

type Lawyer struct {
	Lawyer_name string `json:"lawyerName"`
	Lawyer_id   string `json:"lawyerId"`
	Lawyer_rate int32  `json:"lawyerRate"`
	Lawyer_type string `json:"lawyerType"`
}

type Client struct {
	client_id     string `json:"clientId"`
	client_name   string `json:"clientName"`
	client_budget int32  `json:"clientBudget"`
	client_email  string `json:"clientEmail"`
}

func main() {

	server := echo.New()

	server.Use(middleware.Logger())
	server.Use(middleware.Recover())

	server.GET("/lawyers", getLawyers)
	server.GET("/clients", getClients)
	server.GET("/lawyers/:id", getLawyer)

	server.Logger.Fatal(server.Start(":1323"))
	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func getLawyer(server echo.Context) error {
	fmt.Println("Running...")

	id := server.Param("id")

	var databaseURL = "postgresql://postgres:HovaSihvR070GCJf@db.dentosxmnrtcqboongyz.supabase.co:5432/postgres"
	dbpool, err := pgxpool.New(context.Background(), databaseURL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()
	var lawyer Lawyer

	err = dbpool.QueryRow(context.Background(), "select * from lawyer where id=$1", id).Scan(&lawyer.Lawyer_id, &lawyer.Lawyer_name, &lawyer.Lawyer_rate, &lawyer.Lawyer_type)

	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	return server.JSON(http.StatusOK, lawyer)

}

func getLawyers(server echo.Context) error {
	fmt.Println("Running...")
	var databaseURL = "postgresql://postgres:HovaSihvR070GCJf@db.dentosxmnrtcqboongyz.supabase.co:5432/postgres"
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

		err = rows.Scan(&lawyer.Lawyer_id, &lawyer.Lawyer_name, &lawyer.Lawyer_rate, &lawyer.Lawyer_type)
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
	var databaseURL = "postgresql://postgres:HovaSihvR070GCJf@db.dentosxmnrtcqboongyz.supabase.co:5432/postgres"
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

		err = rows.Scan(&client.client_id, &client.client_name, &client.client_budget, &client.client_email)
		if err != nil {
			log.Fatal("error while iterating dataset")
		}

		clientList = append(clientList, client)

	}
	for _, cl := range clientList {
		fmt.Printf("%s, %s, %d, %s", cl.client_id, cl.client_name, cl.client_budget, cl.client_email)
	}
	return server.String(http.StatusOK, clientList[0].client_name)

}
