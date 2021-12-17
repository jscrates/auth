package main

import (
	"auth/config"
	"auth/handlers"

	"context"
	"flag"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/helmet/v2"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// A MongoInstace contains the Mongo client & database objects
type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

type User struct {
	ID       string `json:"id,omitempty" bson:"_id,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	mg       MongoInstance
	dbName   = config.Get("DB_NAME")
	mongoURI = config.Get("MONGO_URI") + dbName
	prod     = flag.Bool("prod", false, "Enable prefork in Production")
)

func main() {
	// Parse command-line flags
	flag.Parse()

	// Connect with MongoDB
	if err := Connect(); err != nil {
		log.Fatal(err)
	}

	// Initialize a new Fiber app
	app := fiber.New(fiber.Config{
		Prefork: *prod,
	})

	// Middleware
	app.Use(compress.New())
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(helmet.New())
	app.Use(etag.New())
	app.Use(cors.New())

	// Group /v1 endpoint.
	v1 := app.Group("/v1")

	v1.Get("/register", handlers.Register)
	v1.Get("/login", handlers.Login)

	// Handle other 404 routes
	app.Use(handlers.NotFound)

	// Configure port to listen on
	port := config.Get("PORT")

	if config.Get("PORT") == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}

// Connect configures the MongoDB client and initializes the database connection.
// Source: https://www.mongodb.com/blog/post/quick-start-golang--mongodb--starting-and-setup
func Connect() error {
	client, clientError := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	if clientError != nil {
		return clientError
	}

	connectionError := client.Connect(ctx)
	db := client.Database(dbName)

	if connectionError != nil {
		return clientError
	}

	mg = MongoInstance{
		Client: client,
		Db:     db,
	}

	return nil
}
