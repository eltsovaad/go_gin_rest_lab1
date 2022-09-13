// @title Albums API
// @version 1.0
// @description Swagger API for Golang Course Project.
// @termsOfService http://swagger.io/terms/

// @contact.name eltsova_ad
// @contact.email eltsova.ad@gmail.com

// @license.name MIT

// Album model info
// @Description Info about albums
// @Description with its title, artist and my review
package main

import (
	"fmt"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "lab1/docs"
)

type Album struct {
	//ID in data base
	ID uint64 `swaggerignore:"true"`
	//Title of an album
	Title string `form:"title" json:"title"`
	//Artist (band's) name
	Artist string `form:"artist" json:"artist"`
	//My review for the album
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

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/albums", getJsonAlbums)
	r.GET("/albums/:id", getJsonAlbumByID)
	r.POST("/albums", postAlbums)
	r.GET("/albums/new", albumNewGetHandler)
	r.POST("/albums/new", albumNewPostHandler)
	r.GET("/welcome", albumIndexHandler)
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

// getJsonAlbums godoc
// @Summary      Shows Albums as JSON
// @Description  Shows all albums
// @Produce      json
// @Success      200  {object}  Album
// @Failure      500  {int}  http.StatusInternalServerError
// @Router       /albums [get]
func getJsonAlbums(c *gin.Context) {
	db := c.Value("database").(*gorm.DB)
	var albums []Album
	if err := db.Find(&albums).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.IndentedJSON(http.StatusOK, albums)
}

// getJsonAlbums godoc
// @Summary      Shows Albums as JSON
// @Description  get Album by ID
// @Produce      json
// @Param        id query uint false "ID in DB"
// @Success      200  {object}  Album
// @Failure      404  {int}  http.StatusNotFound
// @Router       /albums/{id} [get]
func getJsonAlbumByID(c *gin.Context) {
	db := c.Value("database").(*gorm.DB)
	var album Album

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Can't parse ID"})
		return
	}

	rez := db.Where(&Album{ID: id}).First(&album)
	if rez.RowsAffected == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
		return
	}
	if rez.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error occurred"})
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

// postAlbums godoc
// @Summary      Adds album to the DB
// @Description  post json with album
// @Accept       json
// @Produce      json
// @Param        title body string true "Title of an album"
// @Param        artist body string true "Artist"
// @Param        review body float32 true "My review mark"
// @Success      200  {object}  Album
// @Failure      404  {int}  http.StatusNotFound
// @Failure      500  {int}  http.StatusInternalServerError
// @Router       /albums [post]
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

// @BasePath docs/v1
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
