package main

import (
	"context"
	"fmt"
	"log"

	// "net/url"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoInstance

const dbName = "fiber-hrms"

var mongoURI = "mongodb+srv://hrms.rloww82.mongodb.net/?retryWrites=true&w=majority"

type Employee struct {
	ID     string  `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string  `json:"name"`
	Salary float64 `json:"salary"`
	Age    float64 `json:"age"`
}

func Connect() error {

	credential := options.Credential{
		Username: "Dhruvisha01",
		Password: "Moose_snow@2021",
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI).SetAuth(credential))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	db := client.Database(dbName)

	if err != nil {
		fmt.Printf("Error in connection!")
		return err
	}

	mg = MongoInstance{
		Client: client,
		Db:     db,
	}
	return nil
}

func main() {

	fmt.Println(mongoURI)

	if err := Connect(); err != nil {
		fmt.Printf("Error part 2")
		log.Fatal(err)
	}
	app := fiber.New()

	app.Get("/employee", func(c *fiber.Ctx) error {

		query := bson.D{{}}

		cursor, err := mg.Db.Collection("employees").Find(c.Context(), query)

		if err != nil {
			fmt.Printf("Error part 3")
			return c.Status(500).SendString(err.Error())
		}

		var employees []Employee = make([]Employee, 0)

		if err := cursor.All(c.Context(), &employees); err != nil {
			fmt.Printf("Error 4")
			return c.Status(500).SendString(err.Error())
		}

		return c.JSON(employees)
	})

	app.Post("/employee", func(c *fiber.Ctx) error {
		collection := mg.Db.Collection("employees")

		employee := new(Employee)

		if err := c.BodyParser(employee); err != nil {
			fmt.Printf("Error 5")
			return c.Status(400).SendString(err.Error())
		}

		employee.ID = ""

		insertionResult, err := collection.InsertOne(c.Context(), employee)

		if err != nil {
			fmt.Printf("Error 6")
			return c.Status(500).SendString(err.Error())
		}

		filter := bson.D{{Key: "_id", Value: insertionResult.InsertedID}}
		createdRecord := collection.FindOne(c.Context(), filter)

		createdEmployee := &Employee{}

		createdRecord.Decode(createdEmployee)

		return c.Status(201).JSON(createdEmployee)
	})

	app.Put("/employee/:id", func(c *fiber.Ctx) error {
		idParam := c.Params("id")

		employeeID, err := primitive.ObjectIDFromHex(idParam)

		if err != nil {
			fmt.Printf("Error 7")
			return c.SendStatus(400)
		}

		employee := new(Employee)

		if err := c.BodyParser(employee); err != nil {
			fmt.Printf("Error 8")
			return c.Status(400).SendString(err.Error())
		}

		query := bson.D{{Key: "_id", Value: employeeID}}

		update := bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{Key: "name", Value: employee.Name},
					{Key: "age", Value: employee.Age},
					{Key: "salary", Value: employee.Salary},
				},
			},
		}

		err = mg.Db.Collection("employees").FindOneAndUpdate(c.Context(), query, update).Err()

		if err != nil {
			if err == mongo.ErrNoDocuments {
				fmt.Printf("Error 9")
				return c.SendStatus(400)
			}
			return c.SendStatus(500)
		}

		employee.ID = idParam

		return c.Status(200).JSON(employee)

	})
	app.Delete("/employee/:id", func(c *fiber.Ctx) error {
		employeeID, err := primitive.ObjectIDFromHex(c.Params("id"))

		if err != nil {
			fmt.Printf("Error 10")
			return c.SendStatus(400)
		}

		query := bson.D{{Key: "_id", Value: employeeID}}

		result, err := mg.Db.Collection("employees").DeleteOne(c.Context(), &query)

		if err != nil {
			fmt.Printf("Error 11")
			return c.SendStatus(500)
		}

		if result.DeletedCount < 1 {
			return c.SendStatus(404)
		}

		return c.Status(200).JSON("record deleted")

	})

	log.Fatal(app.Listen(":3000"))
}
