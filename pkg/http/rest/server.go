package rest

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/syfun/bgweb/pkg/db"
)

// Server wraps http.
type Server struct {
	db     *db.Client
	server *http.Server
}

// Run rest server.
func Run(addr string, db *db.Client) {
	router := gin.Default()
	s := &Server{db: db}
	s.route(router)
	s.server = &http.Server{Addr: addr, Handler: router}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func (s *Server) route(router *gin.Engine) {
	router.GET("/api/items/", List(s))
	router.GET("/api/items/:key/", Get(s))
	router.POST("/api/items/", Set(s))
	router.DELETE("/api/items/:key/", Delete(s))
}

type listQuery struct {
	Search   string `form:"search"`
	Page     uint   `form:"page,default=1"`
	PageSize uint   `form:"page_size,default=10"`
}

// List values.
func List(s *Server) func(*gin.Context) {
	return func(c *gin.Context) {
		var query listQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		// if query.Page == 0 {
		// 	query.Page = 1
		// }
		// if query.PageSize == 0 {
		// 	query.PageSize = 10
		// }
		skip := (query.Page - 1) * query.PageSize
		vals, total, err := s.db.List(query.Search, skip, query.PageSize)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"total": total,
			"data":  vals,
		})
	}
}

type setRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// Set value
func Set(s *Server) func(*gin.Context) {
	return func(c *gin.Context) {
		var req setRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := s.db.Set(req.Key, req.Value); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.String(204, "")
	}
}

// Get value
func Get(s *Server) func(*gin.Context) {
	return func(c *gin.Context) {
		key := c.Param("key")
		val, err := s.db.Get(key)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"key": key, "value": val})
	}
}

// Delete value
func Delete(s *Server) func(*gin.Context) {
	return func(c *gin.Context) {
		key := c.Param("key")
		if err := s.db.Delete(key); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(204, "")
	}
}
