package main

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Album struct {
	ID     uint64
	Title  string  `form:"title" json:"title"`
	Artist string  `form:"artist" json:"artist"`
	Review float32 `form:"review" json:"review"`
}

func setupDatabase(db *gorm.DB) error {
	err := db.AutoMigrate(
		&Album{},
	)
	if err != nil {
		return fmt.Errorf("Error migrating database: %s", err)
	}
	return nil
}

func setupRouter(r *gin.Engine, db *gorm.DB) {
	r.Static("/static", "./static/")
	r.LoadHTMLGlob("templates/**/*.html")
	r.Use(connectDatabase(db))
	r.GET("/albums", getJsonAlbums)
	r.GET("/albums/:id", getJsonAlbumByID)
	r.POST("/albums", postAlbums)
	r.GET("/albums/new", albumNewGetHandler)
	r.POST("/albums/new", albumNewPostHandler)
	r.GET("/welcome/", albumIndexHandler)
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/welcome/")
	})
}

// Middleware to connect the database for each request that uses this
// middleware.
func connectDatabase(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("database", db)
	}
}

func albumIndexHandler(c *gin.Context) {
	db := c.Value("database").(*gorm.DB)
	var albums []Album
	if err := db.Find(&albums).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	//db.Where("1 = 1").Delete(&Album{})
	c.HTML(http.StatusOK, "albums/index.html", gin.H{"albums": albums})
}

func getJsonAlbums(c *gin.Context) {
	db := c.Value("database").(*gorm.DB)
	var albums []Album
	if err := db.Find(&albums).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.IndentedJSON(http.StatusOK, albums)
}

func getJsonAlbumByID(c *gin.Context) {
	db := c.Value("database").(*gorm.DB)
	var album Album

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Can't parse ID"})
		return
	}

	db.Where(&Album{ID: id}).First(&album)
	if &album == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, album)
	return
}

func albumNewPostHandler(c *gin.Context) {
	album := &Album{}
	if err := c.Bind(album); err != nil {
		// Note: if there's a bind error, Gin will call
		// c.AbortWithError. We just need to return here.
		return
	}
	// FIXME: There's a better way to do this validation!
	if album.Title == "" || album.Artist == "" || album.Review == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	db := c.Value("database").(*gorm.DB)
	if err := db.Create(&album).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Redirect(http.StatusFound, "/welcome/")
}

func albumNewGetHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "albums/new.html", gin.H{})
}

func postAlbums(c *gin.Context) {
	var newAlbum Album
	fmt.Printf("\nWe are in post Alb\n")
	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.Bind(&newAlbum); err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "can't bind"})
		return
	}
	fmt.Printf("\nAlb:%v\n", newAlbum)
	db := c.Value("database").(*gorm.DB)
	if err := db.Create(&newAlbum).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

func main() {
	db, err := gorm.Open(sqlite.Open("lab1.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}
	err = setupDatabase(db)
	if err != nil {
		log.Fatalf("Database setup error: %s", err)
	}
	r := gin.Default()
	setupRouter(r, db)
	err = r.Run("localhost:8080")
	if err != nil {
		log.Fatalf("gin Run error: %s", err)
	}
}
