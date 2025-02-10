package main

import (
	"database/sql"
	"log"
	"simplebank/api"
	db "simplebank/db/sqlc"
	"simplebank/db/utils"

	_ "github.com/lib/pq"
)


func main() {
    config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config",err)
	}
	conn,err := sql.Open(config.DBDriver,config.DBSource)
	if err != nil {
		log.Fatal("cant connect to the db",err)
	}
   store := db.NewStore(conn) //making connection to the database
   server := api.NewServer(store) //creating a new server

   err = server.Start(config.ServerAddress) //starting the server
   if err != nil {
	log.Fatal("cannot start the server",err)

   }



}