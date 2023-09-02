package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Define a struct for the "mahasiswa" collection.
type Mahasiswa struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	NIM   string             `bson:"nim"`
	Name  string             `bson:"name"`
	Major string             `bson:"major"`
}

func main() {
	// Set up the MongoDB client options.
	clientOptions := options.Client().ApplyURI("mongodb+srv://ucil:ucilucil@golaeng.ingbfjg.mongodb.net/?retryWrites=true&w=majority")

	// Create a context with a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to the MongoDB Atlas cluster.
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Ping the database to check if the connection is successful.
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB Atlas!")

	// Access the "mahasiswa" collection.
	mahasiswaCollection := client.Database("mahasiswaDB").Collection("mahasiswa")

	// Create a new router using the Gorilla Mux router.
	r := mux.NewRouter()

	// Define API endpoints.
	r.HandleFunc("/mahasiswa", getAllMahasiswa(mahasiswaCollection)).Methods("GET")
	r.HandleFunc("/mahasiswa/{id}", getMahasiswa(mahasiswaCollection)).Methods("GET")
	r.HandleFunc("/mahasiswa", createMahasiswa(mahasiswaCollection)).Methods("POST")
	r.HandleFunc("/mahasiswa/{id}", updateMahasiswa(mahasiswaCollection)).Methods("PUT")
	r.HandleFunc("/mahasiswa/{id}", deleteMahasiswa(mahasiswaCollection)).Methods("DELETE")

	// Start the HTTP server.
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":3939", nil))

	// Close the MongoDB connection when done.
	err = client.Disconnect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Disconnected from MongoDB Atlas.")
}

func getAllMahasiswa(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var mahasiswa []Mahasiswa
		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var m Mahasiswa
			if err := cursor.Decode(&m); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			mahasiswa = append(mahasiswa, m)
		}

		if err := cursor.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(mahasiswa); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func getMahasiswa(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		vars := mux.Vars(r)
		id, err := primitive.ObjectIDFromHex(vars["id"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var m Mahasiswa
		err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if err := json.NewEncoder(w).Encode(m); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func createMahasiswa(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var m Mahasiswa
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := collection.InsertOne(ctx, m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		m.ID = result.InsertedID.(primitive.ObjectID)
		if err := json.NewEncoder(w).Encode(m); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func updateMahasiswa(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		vars := mux.Vars(r)
		id, err := primitive.ObjectIDFromHex(vars["id"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var m Mahasiswa
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		update := bson.M{"$set": m}
		_, err = collection.UpdateOne(ctx, bson.M{"_id": id}, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(m); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func deleteMahasiswa(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		vars := mux.Vars(r)
		nim := vars["nim"] // Get the NIM from the URL parameter.

		// Create a filter to match the document with the provided NIM.
		filter := bson.M{"nim": nim}

		result, err := collection.DeleteOne(ctx, filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if result.DeletedCount == 0 {
			http.Error(w, "No document found to delete", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
