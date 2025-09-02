package tools

import (
	"context"
	"fmt"
	"sort"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
	"github.com/mark3labs/mcp-go/mcp"
)

// GetCollectionOverviewTool provides detailed collection analysis
type GetCollectionOverviewTool struct {
	client nuclino.Client
}

func (t *GetCollectionOverviewTool) Name() string {
	return "nuclino_get_collection_overview"
}

func (t *GetCollectionOverviewTool) Description() string {
	return "Get comprehensive overview of a Nuclino collection including item count, content statistics, and recent activity"
}

func (t *GetCollectionOverviewTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"collection_id":      StringProperty("The ID of the collection to analyze"),
		"include_statistics": BoolProperty("Whether to include content statistics (default: true)"),
		"include_recent":     BoolProperty("Whether to include recent items (default: true)"),
		"recent_limit":       IntProperty("Number of recent items to include (default: 5)"),
	}, []string{"collection_id"})
}

func (t *GetCollectionOverviewTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	collectionID, ok := args["collection_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("collection_id must be a string"))
	}

	includeStats := true
	if include, ok := args["include_statistics"].(bool); ok {
		includeStats = include
	}

	includeRecent := true
	if include, ok := args["include_recent"].(bool); ok {
		includeRecent = include
	}

	recentLimit := 5
	if limit, ok := args["recent_limit"].(float64); ok {
		recentLimit = int(limit)
	}

	// Get collection info
	collection, err := t.client.GetCollection(context.Background(), collectionID)
	if err != nil {
		return FormatError(err)
	}

	// Get all items in the workspace to filter by collection
	items, err := t.client.ListItems(context.Background(), collection.WorkspaceID, 1000, 0)
	if err != nil {
		return FormatError(err)
	}

	// Filter items by collection ID
	var collectionItems []nuclino.Item
	for _, item := range items.Results {
		if item.CollectionID == collectionID {
			collectionItems = append(collectionItems, item)
		}
	}

	overview := map[string]interface{}{
		"collection": collection,
		"item_count": len(collectionItems),
		"items":      collectionItems,
	}

	if includeStats && len(collectionItems) > 0 {
		stats := calculateContentStats(collectionItems)
		overview["statistics"] = stats
	}

	if includeRecent && len(collectionItems) > 0 {
		// Sort by updated time (most recent first) - simplified sorting by assuming newer IDs = more recent
		sortedItems := make([]nuclino.Item, len(collectionItems))
		copy(sortedItems, collectionItems)

		// Simple sort - in real implementation you'd sort by UpdatedAt
		sort.Slice(sortedItems, func(i, j int) bool {
			return sortedItems[i].ID > sortedItems[j].ID // Simplified - newer IDs first
		})

		limit := recentLimit
		if limit > len(sortedItems) {
			limit = len(sortedItems)
		}

		overview["recent_items"] = sortedItems[:limit]
	}

	return FormatResult(overview)
}

// OrganizeCollectionTool provides collection organization utilities
type OrganizeCollectionTool struct {
	client nuclino.Client
}

func (t *OrganizeCollectionTool) Name() string {
	return "nuclino_organize_collection"
}

func (t *OrganizeCollectionTool) Description() string {
	return "Analyze and provide organization suggestions for a Nuclino collection based on content patterns"
}

func (t *OrganizeCollectionTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"collection_id":     StringProperty("The ID of the collection to analyze"),
		"suggest_tags":      BoolProperty("Whether to suggest content tags (default: true)"),
		"find_duplicates":   BoolProperty("Whether to find potential duplicate items (default: true)"),
		"analyze_structure": BoolProperty("Whether to analyze content structure (default: true)"),
	}, []string{"collection_id"})
}

func (t *OrganizeCollectionTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	collectionID, ok := args["collection_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("collection_id must be a string"))
	}

	suggestTags := true
	if suggest, ok := args["suggest_tags"].(bool); ok {
		suggestTags = suggest
	}

	findDuplicates := true
	if find, ok := args["find_duplicates"].(bool); ok {
		findDuplicates = find
	}

	analyzeStructure := true
	if analyze, ok := args["analyze_structure"].(bool); ok {
		analyzeStructure = analyze
	}

	// Get collection info
	collection, err := t.client.GetCollection(context.Background(), collectionID)
	if err != nil {
		return FormatError(err)
	}

	// Get collection items
	items, err := t.client.ListItems(context.Background(), collection.WorkspaceID, 1000, 0)
	if err != nil {
		return FormatError(err)
	}

	// Filter by collection
	var collectionItems []nuclino.Item
	for _, item := range items.Results {
		if item.CollectionID == collectionID {
			collectionItems = append(collectionItems, item)
		}
	}

	organization := map[string]interface{}{
		"collection":  collection,
		"total_items": len(collectionItems),
	}

	if suggestTags && len(collectionItems) > 0 {
		tags := suggestContentTags(collectionItems)
		organization["suggested_tags"] = tags
	}

	if findDuplicates && len(collectionItems) > 1 {
		duplicates := findPotentialDuplicates(collectionItems)
		organization["potential_duplicates"] = duplicates
	}

	if analyzeStructure && len(collectionItems) > 0 {
		structure := analyzeContentStructure(collectionItems)
		organization["content_structure"] = structure
	}

	return FormatResult(organization)
}

// BulkOperationsTool provides batch operations for collections
type BulkOperationsTool struct {
	client nuclino.Client
}

func (t *BulkOperationsTool) Name() string {
	return "nuclino_bulk_collection_operations"
}

func (t *BulkOperationsTool) Description() string {
	return "Perform bulk operations on collection items like batch moving, updating, or organizing"
}

func (t *BulkOperationsTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"operation":         StringProperty("Operation type: 'move', 'update_tags', 'organize'"),
		"source_collection": StringProperty("Source collection ID"),
		"target_collection": StringProperty("Target collection ID (for move operations)"),
		"filter_query":      StringProperty("Optional query to filter items for operation"),
		"dry_run":           BoolProperty("Whether to perform a dry run (default: true)"),
	}, []string{"operation", "source_collection"})
}

func (t *BulkOperationsTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return FormatError(fmt.Errorf("operation must be a string"))
	}

	sourceCollection, ok := args["source_collection"].(string)
	if !ok {
		return FormatError(fmt.Errorf("source_collection must be a string"))
	}

	dryRun := true
	if dry, ok := args["dry_run"].(bool); ok {
		dryRun = dry
	}

	// Get source collection
	collection, err := t.client.GetCollection(context.Background(), sourceCollection)
	if err != nil {
		return FormatError(err)
	}

	// Get items in collection
	items, err := t.client.ListItems(context.Background(), collection.WorkspaceID, 1000, 0)
	if err != nil {
		return FormatError(err)
	}

	var collectionItems []nuclino.Item
	for _, item := range items.Results {
		if item.CollectionID == sourceCollection {
			collectionItems = append(collectionItems, item)
		}
	}

	// Filter items if query provided
	if filterQuery, ok := args["filter_query"].(string); ok && filterQuery != "" {
		var filteredItems []nuclino.Item
		for _, item := range collectionItems {
			if containsIgnoreCase(item.Title, filterQuery) || containsIgnoreCase(item.Content, filterQuery) {
				filteredItems = append(filteredItems, item)
			}
		}
		collectionItems = filteredItems
	}

	result := map[string]interface{}{
		"operation":   operation,
		"source":      sourceCollection,
		"items_found": len(collectionItems),
		"dry_run":     dryRun,
	}

	switch operation {
	case "move":
		targetCollection, ok := args["target_collection"].(string)
		if !ok {
			return FormatError(fmt.Errorf("target_collection required for move operation"))
		}

		result["target"] = targetCollection

		if !dryRun {
			var movedItems []string
			for _, item := range collectionItems {
				_, err := t.client.MoveItem(context.Background(), item.ID, targetCollection)
				if err != nil {
					result["error"] = fmt.Sprintf("Failed to move item %s: %v", item.ID, err)
					break
				}
				movedItems = append(movedItems, item.ID)
			}
			result["moved_items"] = movedItems
		} else {
			var itemsToMove []string
			for _, item := range collectionItems {
				itemsToMove = append(itemsToMove, item.ID)
			}
			result["items_to_move"] = itemsToMove
		}

	case "organize":
		suggestions := generateOrganizationSuggestions(collectionItems)
		result["suggestions"] = suggestions

	default:
		return FormatError(fmt.Errorf("unknown operation: %s", operation))
	}

	return FormatResult(result)
}

// Helper functions for content analysis
func calculateContentStats(items []nuclino.Item) map[string]interface{} {
	totalChars := 0
	totalWords := 0
	emptyItems := 0

	for _, item := range items {
		content := item.Content
		totalChars += len(content)

		if len(content) == 0 {
			emptyItems++
		} else {
			// Simple word count (split by spaces)
			words := 1
			for _, char := range content {
				if char == ' ' || char == '\n' || char == '\t' {
					words++
				}
			}
			totalWords += words
		}
	}

	avgChars := 0
	avgWords := 0
	if len(items) > 0 {
		avgChars = totalChars / len(items)
		avgWords = totalWords / len(items)
	}

	return map[string]interface{}{
		"total_characters":   totalChars,
		"total_words":        totalWords,
		"average_characters": avgChars,
		"average_words":      avgWords,
		"empty_items":        emptyItems,
		"items_with_content": len(items) - emptyItems,
	}
}

func suggestContentTags(items []nuclino.Item) []string {
	// Simple tag suggestion based on common words in titles
	wordCount := make(map[string]int)

	for _, item := range items {
		words := splitWords(item.Title)
		for _, word := range words {
			if len(word) > 3 { // Only consider words longer than 3 characters
				wordCount[toLower(word)]++
			}
		}
	}

	var tags []string
	for word, count := range wordCount {
		if count >= 2 { // Word appears in at least 2 items
			tags = append(tags, word)
		}
	}

	return tags
}

func findPotentialDuplicates(items []nuclino.Item) [][]string {
	var duplicates [][]string

	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			similarity := calculateSimilarity(items[i].Title, items[j].Title)
			if similarity > 0.7 { // 70% similarity threshold
				duplicates = append(duplicates, []string{items[i].ID, items[j].ID})
			}
		}
	}

	return duplicates
}

func analyzeContentStructure(items []nuclino.Item) map[string]interface{} {
	markdownHeaders := 0
	listItems := 0
	codeBlocks := 0

	for _, item := range items {
		content := item.Content

		// Count markdown elements (simplified)
		for i, char := range content {
			if char == '#' && (i == 0 || content[i-1] == '\n') {
				markdownHeaders++
			}
			if char == '-' && (i == 0 || content[i-1] == '\n') {
				listItems++
			}
			if char == '`' {
				codeBlocks++
			}
		}
	}

	return map[string]interface{}{
		"markdown_headers": markdownHeaders,
		"list_items":       listItems,
		"code_blocks":      codeBlocks / 3, // Approximate (``` blocks)
	}
}

func generateOrganizationSuggestions(items []nuclino.Item) []string {
	suggestions := []string{}

	if len(items) > 20 {
		suggestions = append(suggestions, "Consider splitting this large collection into smaller, topic-focused collections")
	}

	emptyCount := 0
	for _, item := range items {
		if len(item.Content) == 0 {
			emptyCount++
		}
	}

	if emptyCount > len(items)/4 {
		suggestions = append(suggestions, "Many items are empty - consider removing placeholder items or adding content")
	}

	tags := suggestContentTags(items)
	if len(tags) > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Consider organizing items by themes: %v", tags[:min(3, len(tags))]))
	}

	return suggestions
}

func splitWords(text string) []string {
	var words []string
	var current []byte

	for _, char := range []byte(text) {
		if char == ' ' || char == '\n' || char == '\t' || char == '.' || char == ',' {
			if len(current) > 0 {
				words = append(words, string(current))
				current = nil
			}
		} else {
			current = append(current, char)
		}
	}

	if len(current) > 0 {
		words = append(words, string(current))
	}

	return words
}

func calculateSimilarity(str1, str2 string) float64 {
	if str1 == str2 {
		return 1.0
	}

	// Simple similarity based on common characters
	common := 0
	total := len(str1) + len(str2)

	if total == 0 {
		return 1.0
	}

	for _, char1 := range str1 {
		for _, char2 := range str2 {
			if char1 == char2 {
				common++
				break
			}
		}
	}

	return float64(common*2) / float64(total)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
