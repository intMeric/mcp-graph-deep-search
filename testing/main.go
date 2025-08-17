package main

import (
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("No .env file found or error loading it: %v", err)
	}
	id := "www.liberation.fr/economie/medias/polemique-autour-du-financement-de-blast-le-nouveau-media-de-denis-robert-20210311_X4AAQZKAMNFE7FWYRSCBK3USQY/"
	hash := sha256.Sum256([]byte(id))
	var objectIDBytes [12]byte
	copy(objectIDBytes[:], hash[:12])
	r := primitive.ObjectID(objectIDBytes)
	fmt.Println(r)

	// analyzeWebpageUseCase, err := search_and_analyze.NewSearchAndAnalyzeUseCase()
	// if err != nil {
	// 	log.Fatalf("%v", err)
	// }
	// r, err := analyzeWebpageUseCase.Execute(context.TODO(), &search_and_analyze.SearchAndAnalyzeRequest{
	// 	Query:      "Ground News media bias news aggregation",
	// 	MaxResults: 15,
	// })
	// fmt.Printf("Search and analyze result: %v\n", r)
	// if err != nil {
	// 	log.Fatalf("Error executing search and analyze use case: %v", err)
	// }

	// db, err3 := database.NewMongoDatabase("documents") // Default collection for document search
	// if err3 != nil {
	// 	log.Fatalf("Error initializing database: %v", err3)
	// }

	// searchUseCase := search_document.NewSearchDocumentUseCase(db)
	// result, err := searchUseCase.Execute(context.TODO(), &node.Node{
	// 	ID:       "ground.news/subscribe",
	// 	Location: "webpage",
	// })
	// if err != nil {
	// 	log.Fatalf("Error executing search document use case: %v", err)
	// }

	// log.Printf("Search result: %v", result)
}
