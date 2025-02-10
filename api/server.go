package api

import (
	db "simplebank/db/sqlc"

	"github.com/gin-gonic/gin"
)

//server serves Http requests for our banking service
type Server struct {
   store *db.Store
   router *gin.Engine
//router help us each api request to the correct handler for processing
}


//NewServer creates a new HTTP server and set up routing
func NewServer(store *db.Store) *Server {
   server := &Server{store: store}
   router := gin.Default()
   router.POST("/accounts", server.createAccount)
   router.GET("/accounts/:id", server.getAccount)
   

   server.router = router  
   return server
}

//Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
  return server.router.Run(address)
}

func errResponse(err error) gin.H {
   return gin.H{"error": err.Error()}
}
