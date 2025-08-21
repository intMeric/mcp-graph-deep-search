package main

import (
	"context"
	"fmt"
	"mgds/src/pkg/search_engine"
	"time"
)

func main() {
	fmt.Println("🔍 Testing Search Engine with SearchXNG...")

	// Configure client for your local SearchXNG instance
	config := &search_engine.Config{
		BaseURL:           "http://localhost:8888",
		Timeout:           30 * time.Second,
		UserAgent:         "mgds-test/1.0",
		DefaultLanguage:   "en",
		DefaultSafeSearch: 1,
		MaxResults:        20,
		DefaultTimeRange:  search_engine.TimeRangeShort,
	}

	// Create SearchXNG client
	client := search_engine.NewSearXNG(config)
	defer client.Close()

	ctx := context.Background()

	// Test 1: Health check
	fmt.Println("\n💊 Testing SearchXNG health...")
	if err := client.IsHealthy(ctx); err != nil {
		fmt.Printf("❌ SearchXNG health check failed: %v\n", err)
		fmt.Println("Make sure your SearchXNG instance is running on http://localhost:8888")
		return
	}
	fmt.Println("✅ SearchXNG is healthy!")

	// Test 2: Search for "Mersen" with short time range (general + news)
	fmt.Println("\n🔍 Searching for 'Mersen' (short time, general + news)...")

	req := &search_engine.SearchRequest{
		Query:      "Mersen",
		Categories: []search_engine.Category{search_engine.CategoryGeneral, search_engine.CategoryNews},
		TimeRange:  search_engine.TimeRangeShort,
		Language:   "en",
		SafeSearch: 1,
		PageNo:     1,
		MaxResults: 10,
	}

	startTime := time.Now()
	resp, err := client.Search(ctx, req)
	if err != nil {
		fmt.Printf("❌ Search failed: %v\n", err)
		return
	}

	searchDuration := time.Since(startTime)

	// Display results
	fmt.Printf("✅ Search completed in %v (API reported: %v)\n", searchDuration, resp.TimeElapsed)
	fmt.Printf("📊 Found %d results for query: '%s'\n", len(resp.Results), resp.Query)
	fmt.Printf("📄 Page %d/%d (Next: %v, Prev: %v)\n",
		resp.CurrentPage,
		resp.CurrentPage,
		resp.HasNextPage,
		resp.HasPrevPage)

	if len(resp.Results) == 0 {
		fmt.Println("🤷 No results found. This might be because:")
		fmt.Println("  - Your SearchXNG instance doesn't have search engines configured")
		fmt.Println("  - The time range 'short' is too restrictive")
		fmt.Println("  - Network connectivity issues")
		return
	}

	fmt.Println("\n📋 Results:")
	for i, result := range resp.Results {
		fmt.Printf("\n%d. %s\n", i+1, result.Title)
		fmt.Printf("   🌐 URL: %s\n", result.URL)
		fmt.Printf("   📂 Category: %s\n", result.Category)
		fmt.Printf("   🔧 Engine: %s\n", result.Engine)
		fmt.Printf("   ⭐ Score: %.2f\n", result.Score)

		if result.PublishedDate != nil {
			fmt.Printf("   📅 Published: %s\n", result.PublishedDate.Format("2006-01-02 15:04:05"))
		}

		if len(result.Content) > 0 {
			content := result.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			fmt.Printf("   📝 Content: %s\n", content)
		}

		if result.Metadata != nil {
			if engines, ok := result.Metadata["engines"].([]string); ok && len(engines) > 1 {
				fmt.Printf("   🔍 Also found in: %v\n", engines[1:])
			}
		}
	}

	// Test 3: Simple search test
	fmt.Println("\n\n🚀 Testing simple search...")
	simpleResp, err := client.SearchSimple(ctx, "Mersen stock")
	if err != nil {
		fmt.Printf("❌ Simple search failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Simple search found %d results for 'Mersen stock'\n", len(simpleResp.Results))
	if len(simpleResp.Results) > 0 {
		fmt.Printf("📰 First result: %s\n", simpleResp.Results[0].Title)
		fmt.Printf("🌐 URL: %s\n", simpleResp.Results[0].URL)
	}

	fmt.Println("\n✅ Manual test completed successfully!")
	fmt.Println("🔧 SearchXNG configuration:")
	fmt.Printf("   • Instance: %s\n", config.BaseURL)
	fmt.Printf("   • Timeout: %v\n", config.Timeout)
	fmt.Printf("   • User Agent: %s\n", config.UserAgent)
	fmt.Printf("   • Default Language: %s\n", config.DefaultLanguage)
	fmt.Printf("   • Max Results: %d\n", config.MaxResults)
}

// func main1() {
// 	fmt.Println("🚀 Testing MCP Graph Deep Search with Cayley backend...")

// 	// Create temporary directory for test database
// 	tempDir, err := os.MkdirTemp("", "mgds-test-*")
// 	if err != nil {
// 		log.Fatalf("Failed to create temp dir: %v", err)
// 	}
// 	defer os.RemoveAll(tempDir)

// 	dbPath := filepath.Join(tempDir, "test-graph.db")
// 	fmt.Printf("📁 Using database: %s\n", dbPath)

// 	// Create graph with memory backend (persistent storage will be added later)
// 	config := &graph.Config{
// 		DatabasePath: ":memory:",
// 		DatabaseType: "memory",
// 	}

// 	g := graph.NewCayleyGraph(config)
// 	ctx := context.Background()

// 	// Connect to database
// 	fmt.Println("🔌 Connecting to database...")
// 	if err := g.Connect(ctx); err != nil {
// 		log.Fatalf("Failed to connect: %v", err)
// 	}
// 	defer g.Close()

// 	// Test ping
// 	if err := g.Ping(ctx); err != nil {
// 		log.Fatalf("Ping failed: %v", err)
// 	}
// 	fmt.Println("✅ Database connection successful")

// 	// Test 1: Create and persist nodes
// 	fmt.Println("\n📝 Test 1: Creating nodes...")

// 	node1 := object.NewURLNodeWithDetails("web-1", "OpenAI Website", "Official OpenAI website", "https://openai.com",
// 		"OpenAI", "OpenAI is an AI research company. We develop AI systems that can help humans.")
// 	node2 := object.NewURLNodeWithDetails("web-2", "Anthropic Website", "Official Anthropic website", "https://anthropic.com",
// 		"Anthropic", "Anthropic builds AI systems that are safe, beneficial, and understandable.")
// 	node3 := object.NewURLNodeWithDetails("web-3", "Claude AI", "Claude AI interface", "https://claude.ai",
// 		"Claude", "Claude is an AI assistant created by Anthropic. I'm helpful, harmless, and honest.")

// 	// Set some properties
// 	node1.SetProperty("category", "AI Company")
// 	node2.SetProperty("category", "AI Company")
// 	node3.SetProperty("category", "AI Product")

// 	// Add nodes
// 	nodes := []string{"web-1", "web-2", "web-3"}
// 	for i, node := range []*object.URLNode{
// 		node1.(*object.URLNode),
// 		node2.(*object.URLNode),
// 		node3.(*object.URLNode),
// 	} {
// 		if err := g.AddNode(ctx, node); err != nil {
// 			log.Fatalf("Failed to add node %s: %v", nodes[i], err)
// 		}
// 		fmt.Printf("  ✓ Added node: %s (%s)\n", node.GetDisplayName(), node.GetMgdsId())
// 	}

// 	// Test 2: Create relations
// 	fmt.Println("\n🔗 Test 2: Creating relations...")

// 	relations := []*graph.Relation{
// 		{FromNodeId: "web-1", ToNodeId: "web-3", Label: "mentions", Properties: map[string]any{"context": "competitor"}},
// 		{FromNodeId: "web-2", ToNodeId: "web-3", Label: "creates", Properties: map[string]any{"type": "product"}},
// 		{FromNodeId: "web-1", ToNodeId: "web-2", Label: "competes_with", Properties: map[string]any{"market": "AI"}},
// 	}

// 	for _, rel := range relations {
// 		if err := g.AddRelation(ctx, rel); err != nil {
// 			log.Fatalf("Failed to add relation: %v", err)
// 		}
// 		fmt.Printf("  ✓ Added relation: %s -[%s]-> %s\n", rel.FromNodeId, rel.Label, rel.ToNodeId)
// 	}

// 	// Test 3: Query nodes
// 	fmt.Println("\n🔍 Test 3: Querying nodes...")

// 	// Get all nodes
// 	allNodes, err := g.GetAllNodes(ctx)
// 	if err != nil {
// 		log.Fatalf("Failed to get all nodes: %v", err)
// 	}
// 	fmt.Printf("  📊 Found %d nodes in database\n", len(allNodes))

// 	// Get nodes by type
// 	urlNodes, err := g.FindNodesByType(ctx, object.URLNodeType)
// 	if err != nil {
// 		log.Fatalf("Failed to find URL nodes: %v", err)
// 	}
// 	fmt.Printf("  🌐 Found %d URL nodes\n", len(urlNodes))

// 	// Test individual node retrieval with content verification
// 	retrievedNode, err := g.GetNode(ctx, "web-2")
// 	if err != nil {
// 		log.Fatalf("Failed to retrieve node: %v", err)
// 	}
// 	fmt.Printf("  📄 Retrieved node: %s - %s\n", retrievedNode.GetMgdsId(), retrievedNode.GetDisplayName())

// 	// Verify URLNode-specific fields are persisted
// 	if urlNode, ok := retrievedNode.(*object.URLNode); ok {
// 		fmt.Printf("    🌐 URL: %s\n", urlNode.GetURL())
// 		fmt.Printf("    📰 Title: %s\n", urlNode.GetTitle())
// 		fmt.Printf("    📝 Content: %.50s...\n", urlNode.GetContent())
// 		fmt.Printf("    🔍 Scrapped: %v\n", urlNode.IsScrapped())
// 		if category, exists := urlNode.GetProperty("category"); exists {
// 			fmt.Printf("    🏷️  Category: %s\n", category)
// 		}
// 	}

// 	// Test 4: Query relations
// 	fmt.Println("\n🔗 Test 4: Querying relations...")

// 	// Get outgoing relations from web-1
// 	outgoing, err := g.GetRelations(ctx, "web-1")
// 	if err != nil {
// 		log.Fatalf("Failed to get outgoing relations: %v", err)
// 	}
// 	fmt.Printf("  📤 Node web-1 has %d outgoing relations:\n", len(outgoing))
// 	for _, rel := range outgoing {
// 		fmt.Printf("    → %s -[%s]-> %s\n", rel.FromNodeId, rel.Label, rel.ToNodeId)
// 	}

// 	// Get incoming relations to web-3
// 	incoming, err := g.GetIncomingRelations(ctx, "web-3")
// 	if err != nil {
// 		log.Fatalf("Failed to get incoming relations: %v", err)
// 	}
// 	fmt.Printf("  📥 Node web-3 has %d incoming relations:\n", len(incoming))
// 	for _, rel := range incoming {
// 		fmt.Printf("    → %s -[%s]-> %s\n", rel.FromNodeId, rel.Label, rel.ToNodeId)
// 	}

// 	// Test 5: Graph traversal
// 	fmt.Println("\n🗺️  Test 5: Graph traversal...")

// 	// Get connected nodes
// 	connected, err := g.GetConnectedNodes(ctx, "web-2")
// 	if err != nil {
// 		log.Fatalf("Failed to get connected nodes: %v", err)
// 	}
// 	fmt.Printf("  🔗 Node web-2 is connected to %d nodes:\n", len(connected))
// 	for _, node := range connected {
// 		fmt.Printf("    → %s (%s)\n", node.GetDisplayName(), node.GetMgdsId())
// 	}

// 	// Test path finding
// 	path, err := g.FindPath(ctx, "web-1", "web-3")
// 	if err != nil {
// 		log.Fatalf("Failed to find path: %v", err)
// 	}
// 	if len(path) > 0 {
// 		fmt.Printf("  🛤️  Found direct path from web-1 to web-3: %s\n", path[0].Label)
// 	} else {
// 		fmt.Println("  🛤️  No direct path found from web-1 to web-3")
// 	}

// 	// Test 6: Update operations
// 	fmt.Println("\n✏️  Test 6: Update operations...")

// 	// Update a node
// 	node1Concrete := node1.(*object.URLNode)
// 	node1Concrete.SetTitle("OpenAI - Leading AI Research")
// 	node1Concrete.SetScrapped(true)

// 	if err := g.UpdateNode(ctx, node1Concrete); err != nil {
// 		log.Fatalf("Failed to update node: %v", err)
// 	}
// 	fmt.Println("  ✓ Updated node web-1")

// 	// Verify update with detailed check
// 	updated, err := g.GetNode(ctx, "web-1")
// 	if err != nil {
// 		log.Fatalf("Failed to retrieve updated node: %v", err)
// 	}
// 	fmt.Printf("  📄 Verified update: %s\n", updated.GetDisplayName())

// 	// Verify that URLNode-specific updates were persisted
// 	if updatedURLNode, ok := updated.(*object.URLNode); ok {
// 		fmt.Printf("    📰 Updated title: %s\n", updatedURLNode.GetTitle())
// 		fmt.Printf("    🔍 Scrapped status: %v\n", updatedURLNode.IsScrapped())
// 		fmt.Printf("    🌐 Original URL: %s\n", updatedURLNode.GetURL())
// 		fmt.Printf("    📝 Original content: %.30s...\n", updatedURLNode.GetContent())
// 	}

// 	// Test 7: Cleanup operations
// 	fmt.Println("\n🧹 Test 7: Cleanup operations...")

// 	// Delete a relation
// 	if err := g.DeleteRelation(ctx, "web-1", "web-2", "competes_with"); err != nil {
// 		log.Fatalf("Failed to delete relation: %v", err)
// 	}
// 	fmt.Println("  ✓ Deleted relation: web-1 -[competes_with]-> web-2")

// 	// Verify relation was deleted
// 	outgoingAfterDelete, err := g.GetRelations(ctx, "web-1")
// 	if err != nil {
// 		log.Fatalf("Failed to get relations after delete: %v", err)
// 	}
// 	fmt.Printf("  📊 Node web-1 now has %d outgoing relations\n", len(outgoingAfterDelete))

// 	// Final verification
// 	fmt.Println("\n🔍 Final verification...")
// 	finalCount, err := g.GetAllNodes(ctx)
// 	if err != nil {
// 		log.Fatalf("Failed final node count: %v", err)
// 	}
// 	fmt.Printf("  📊 Final node count: %d\n", len(finalCount))

// 	// Show database size
// 	if stat, err := os.Stat(dbPath); err == nil {
// 		fmt.Printf("  💾 Database directory created: %d bytes\n", stat.Size())
// 	}

// 	fmt.Println("\n✅ All tests completed successfully!")
// 	fmt.Printf("🗄️  Database persisted at: %s\n", dbPath)
// 	fmt.Println("📝 Test demonstrates:")
// 	fmt.Println("   - Cayley graph database integration")
// 	fmt.Println("   - Node CRUD operations")
// 	fmt.Println("   - Relation management")
// 	fmt.Println("   - Graph traversal")
// 	fmt.Println("   - Complete graph API functionality")

// 	fmt.Println("\n⏰ Keeping database alive for 5 seconds for inspection...")
// 	time.Sleep(5 * time.Second)
// }
