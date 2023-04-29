package main

import (
	"flag"
	"log"

	"github.com/yosupo06/library-checker-judge/database"
)

func main() {
	pgHost := flag.String("pghost", "localhost", "postgre host")
	pgUser := flag.String("pguser", "postgres", "postgre user")
	pgPass := flag.String("pgpass", "passwd", "postgre password")
	pgTable := flag.String("pgtable", "librarychecker", "postgre table name")

	flag.Parse()

	// connect db
	db := database.Connect(
		*pgHost,
		"5432",
		*pgTable,
		*pgUser,
		*pgPass,
		false)

	if err := database.RegisterUser(db, "admin", "password", true); err != nil {
		log.Fatalln("failed to add user, admin:", err)
	}
	if err := database.RegisterUser(db, "tester", "password", false); err != nil {
		log.Fatalln("failed to add user, tester:", err)
	}

	log.Println("test user added")
}
