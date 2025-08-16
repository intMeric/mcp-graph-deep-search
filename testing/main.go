package main

import (
	"context"
	"log"
	"mgds/internal/app/use-cases/search_document"
	"mgds/internal/pkg/database"
	"mgds/internal/pkg/node"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("No .env file found or error loading it: %v", err)
	}

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

	db, err3 := database.NewMongoDatabase("documents") // Default collection for document search
	if err3 != nil {
		log.Fatalf("Error initializing database: %v", err3)
	}

	searchUseCase := search_document.NewSearchDocumentUseCase(db)
	result, err := searchUseCase.Execute(context.TODO(), &node.Node{
		ID:       "ground.news/subscribe",
		Location: "webpage",
	})
	if err != nil {
		log.Fatalf("Error executing search document use case: %v", err)
	}

	log.Printf("Search result: %v", result)
}
