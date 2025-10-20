package main

import (
	"fmt"
	"os"

	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "week11-assignment/docs"
)

type Book struct {
	ID     int     `json:"id"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	ISBN   string  `json:"isbn"`
	Year   int     `json:"year"`
	Price  float64 `json:"price"`

	// ฟิลด์ใหม่
	Category      string   `json:"category"`
	OriginalPrice *float64 `json:"original_price,omitempty"`
	Discount      int      `json:"discount"`
	CoverImage    string   `json:"cover_image"`
	Rating        float64  `json:"rating"`
	ReviewsCount  int      `json:"reviews_count"`
	IsNew         bool     `json:"is_new"`
	Pages         *int     `json:"pages,omitempty"`
	Language      string   `json:"language"`
	Publisher     string   `json:"publisher"`
	Description   string   `json:"description"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var db *sql.DB

func initDB() {
	var err error
	host := getEnv("DB_HOST", "localhost")
	name := getEnv("DB_NAME", "bookstore")
	user := getEnv("DB_USER", "bookstore_user")
	password := getEnv("DB_PASSWORD", "your_strong_password")
	port := getEnv("DB_PORT", "5432")

	conSt := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)
	//fmt.Println(conSt)

	db, err = sql.Open("postgres", conSt)
	if err != nil {
		log.Fatal("failed to open database")
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to database")
	}

	log.Print("successfully connect to database")

	// กำหนดจำนวน Connection สูงสุด
	db.SetMaxOpenConns(25)

	// กำหนดจำนวน Idle connection สูงสุด
	db.SetMaxIdleConns(20)

	// กำหนดอายุของ Connection
	db.SetConnMaxLifetime(5 * time.Minute)
}

// @Summary Get all book
// @Description Get details of a books
// @Tags Books
// @Produce  json
// @Success 200  {object}  Book
// @Failure 404  {object}  ErrorResponse
// @Router  /books [get]
func getAllBooks(c *gin.Context) {
	var rows *sql.Rows
	var err error
	categoryQuery := c.Query("category")

	if categoryQuery != "" {
		// ลูกค้าถาม "มีหนังสืออะไรบ้าง"
		rows, err = db.Query(`
            SELECT id, title, author, category, isbn, year, price, created_at, updated_at FROM books WHERE category = $1`, categoryQuery)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer rows.Close() // ต้องปิด rows เสมอ เพื่อคืน Connection กลับ pool

	} else {
		// ลูกค้าถาม "มีหนังสืออะไรบ้าง"
		rows, err = db.Query(`
            SELECT id, title, author, category, isbn, year, price, created_at, updated_at FROM books`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer rows.Close() // ต้องปิด rows เสมอ เพื่อคืน Connection กลับ pool
	}

	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Category, &book.ISBN, &book.Year, &book.Price, &book.CreatedAt, &book.UpdatedAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		books = append(books, book)
	}

	if books == nil {
		books = []Book{}
	}

	c.JSON(http.StatusOK, books)
}

// @Summary get new book
// @Description get new book
// @Tags Books
// @Accept  json
// @Produce  json
// @Success 200 {object} Book
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/new [get]
func getNewBooks(c *gin.Context) {

	rows, err := db.Query(`
        SELECT id, title, author, isbn, year, price, created_at, updated_at
        FROM books
        ORDER BY created_at DESC
        LIMIT 5
    `)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Author,
			&book.ISBN,
			&book.Year,
			&book.Price,
			&book.CreatedAt,
			&book.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books = append(books, book)
	}

	if books == nil {
		books = []Book{}
	}

	c.JSON(http.StatusOK, books)
}

// @Summary Get book by ID
// @Description Get details of a book by its ID
// @Tags Books
// @Produce  json
// @Param   id   path   int  true  "Book ID"
// @Success 200 {object} Book
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/{id} [get]
func getBook(c *gin.Context) {
	id := c.Param("id")
	var book Book

	// QueryRow ใช้เมื่อคาดว่าจะได้ผลลัพธ์ 0 หรือ 1 แถว
	err := db.QueryRow("SELECT id, title, author FROM books WHERE id = $1", id).
		Scan(&book.ID, &book.Title, &book.Author)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, book)
}

// @Summary Create a new book
// @Description Add a new book record to the database
// @Tags Books
// @Accept  json
// @Produce  json
// @Param   book  body  Book  true  "Book data"
// @Success 201 {object} Book
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books [post]
func createBook(c *gin.Context) {
	var newBook Book

	if err := c.ShouldBindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ใช้ RETURNING เพื่อดึงค่าที่ database generate (id, timestamps)
	var id int
	var createdAt, updatedAt time.Time

	err := db.QueryRow(
		`INSERT INTO books (title, author, isbn, year, price)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id, created_at, updated_at`,
		newBook.Title, newBook.Author, newBook.ISBN, newBook.Year, newBook.Price,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newBook.ID = id
	newBook.CreatedAt = createdAt
	newBook.UpdatedAt = updatedAt

	c.JSON(http.StatusCreated, newBook) // ใช้ 201 Created
}

// @Summary Update book by ID
// @Description Update existing book details by its ID
// @Tags Books
// @Accept  json
// @Produce  json
// @Param   id    path  int   true  "Book ID"
// @Param   book  body  Book  true  "Updated book data"
// @Success 200 {object} Book
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/{id} [put]
func updateBook(c *gin.Context) {
	var ID int
	id := c.Param("id")
	var updateBook Book

	if err := c.ShouldBindJSON(&updateBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updatedAt time.Time
	err := db.QueryRow(
		`UPDATE books
         SET title = $1, author = $2, isbn = $3, year = $4, price = $5
         WHERE id = $6
         RETURNING ID,updated_at`,
		updateBook.Title, updateBook.Author, updateBook.ISBN,
		updateBook.Year, updateBook.Price, id,
	).Scan(&ID, &updatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updateBook.ID = ID
	updateBook.UpdatedAt = updatedAt
	c.JSON(http.StatusOK, updateBook)
}

// @Summary Delete book by ID
// @Description Delete a book record from the database by its ID
// @Tags Books
// @Produce  json
// @Param   id  path  int  true  "Book ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/{id} [delete]
func deleteBook(c *gin.Context) {
	id := c.Param("id")

	result, err := db.Exec("DELETE FROM books WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "book deleted successfully"})
}

func getAllCategories(c *gin.Context) {
	rows, err := db.Query("SELECT DISTINCT category FROM books")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		rows.Scan(&category)
		categories = append(categories, category)
	}

	if categories == nil {
		categories = []string{}
	}

	c.JSON(http.StatusOK, categories)
}

func searchwantBook(c *gin.Context) {
	searchQuery := c.Query("keyword")

	if searchQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keyword is required"})
		return
	}

	if searchQuery != "" {
		keyword := "%" + searchQuery + "%"
		rows, err := db.Query(`
            SELECT id, title, author, category, isbn, year, price, created_at, updated_at FROM books WHERE title ILIKE $1  OR author ILIKE $1
       OR category ILIKE $1 OR isbn ILIKE $1`, keyword)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var books []Book
		for rows.Next() {
			var book Book
			err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Category, &book.ISBN, &book.Year, &book.Price, &book.CreatedAt, &book.UpdatedAt)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			books = append(books, book)
		}

		if books == nil {
			books = []Book{}
		}
		c.JSON(http.StatusOK, books)
	}

}

func recommandedBook(c *gin.Context) {
	rows, err := db.Query(`
        SELECT id, title, author, category, isbn, year, price, created_at, updated_at, rating
        FROM books
        ORDER BY rating DESC
        LIMIT 5
    `)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Author,
			&book.Category,
			&book.ISBN,
			&book.Year,
			&book.Price,
			&book.CreatedAt,
			&book.UpdatedAt,
			&book.Rating,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books = append(books, book)
	}

	if books == nil {
		books = []Book{}
	}

	c.JSON(http.StatusOK, books)
}

func discountedbook(c *gin.Context) {

	rows, err := db.Query(`
        SELECT id, title, author, category, isbn, year, price, discount, created_at, updated_at, rating
        FROM books
        ORDER BY discount DESC
        LIMIT 5
    `)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Author,
			&book.Category,
			&book.ISBN,
			&book.Year,
			&book.Price,
			&book.Discount,
			&book.CreatedAt,
			&book.UpdatedAt,
			&book.Rating,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books = append(books, book)
	}

	if books == nil {
		books = []Book{}
	}

	c.JSON(http.StatusOK, books)

}

// @title           Simple API Example
// @version         1.0
// @description     This is a simple example of using Gin with Swagger.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	initDB()
	defer db.Close()

	r := gin.Default()
	r.Use(cors.Default())
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		err := db.Ping()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"message": "unhealty", "error": err})
			return
		}
		c.JSON(200, gin.H{"message": "healthy"})
	})

	api := r.Group("/api/v1")
	{
		api.GET("/books", getAllBooks)
		api.GET("/books/new", getNewBooks)
		api.GET("/books/:id", getBook)
		api.POST("/books", createBook)
		api.PUT("/books/:id", updateBook)
		api.DELETE("/books/:id", deleteBook)

		api.GET("/categories", getAllCategories)
		api.GET("/books/search", searchwantBook)
		api.GET("/books/featured", recommandedBook)
		api.GET("/books/discounted", discountedbook)
	}

	r.Run(":8080")
}
