package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
)

func main() {

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

		err = rows.Scan(&lawyer.Lawyer_id, &lawyer.Lawyer_name, &lawyer.lawyer_rate, &lawyer.Lawyer_type)
		if err != nil {
			log.Fatal("error while iterating dataset")
		}

		lawyerList = append(lawyerList, lawyer)

	}
	for _, la := range lawyerList {
		fmt.Printf("%s, %s, %d, %s", la.Lawyer_id, la.Lawyer_name, la.lawyer_rate, la.Lawyer_type)
	}

}
