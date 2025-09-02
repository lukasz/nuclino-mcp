package nuclino

import "time"

// User represents a Nuclino user
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Team represents a Nuclino team
type Team struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Workspace represents a Nuclino workspace
type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	TeamID    string    `json:"teamId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Collection represents a Nuclino collection
type Collection struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	WorkspaceID string    `json:"workspaceId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Item represents a Nuclino item
type Item struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	CollectionID string    `json:"collectionId"`
	WorkspaceID  string    `json:"workspaceId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	CreatedBy    string    `json:"createdBy"`
	UpdatedBy    string    `json:"updatedBy"`
	URL          string    `json:"url"`
}

// CreateItemRequest represents the request to create a new item
type CreateItemRequest struct {
	Title        string `json:"title" validate:"required"`
	Content      string `json:"content"`
	CollectionID string `json:"collectionId" validate:"required"`
}

// UpdateItemRequest represents the request to update an item
type UpdateItemRequest struct {
	Title        *string `json:"title,omitempty"`
	Content      *string `json:"content,omitempty"`
	CollectionID *string `json:"collectionId,omitempty"`
}

// SearchItemsRequest represents the request for searching items
type SearchItemsRequest struct {
	Query       string `json:"query,omitempty"`
	WorkspaceID string `json:"workspaceId,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

// CreateWorkspaceRequest represents the request to create a workspace
type CreateWorkspaceRequest struct {
	Name   string `json:"name" validate:"required"`
	TeamID string `json:"teamId" validate:"required"`
}

// UpdateWorkspaceRequest represents the request to update a workspace
type UpdateWorkspaceRequest struct {
	Name *string `json:"name,omitempty"`
}

// CreateCollectionRequest represents the request to create a collection
type CreateCollectionRequest struct {
	Title       string `json:"title" validate:"required"`
	WorkspaceID string `json:"workspaceId" validate:"required"`
}

// UpdateCollectionRequest represents the request to update a collection
type UpdateCollectionRequest struct {
	Title *string `json:"title,omitempty"`
}

// Field represents a Nuclino field
type Field struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	WorkspaceID string    `json:"workspaceId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// File represents a Nuclino file
type File struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mimeType"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Results []interface{} `json:"results"`
	Total   int           `json:"total"`
	Limit   int           `json:"limit"`
	Offset  int           `json:"offset"`
}

// ItemsResponse represents the response for items list
type ItemsResponse struct {
	Results []Item `json:"results"`
	Total   int    `json:"total"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
}

// WorkspacesResponse represents the response for workspaces list
type WorkspacesResponse struct {
	Results []Workspace `json:"results"`
	Total   int         `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
}

// CollectionsResponse represents the response for collections list
type CollectionsResponse struct {
	Results []Collection `json:"results"`
	Total   int          `json:"total"`
	Limit   int          `json:"limit"`
	Offset  int          `json:"offset"`
}

// TeamsResponse represents the response for teams list
type TeamsResponse struct {
	Results []Team `json:"results"`
	Total   int    `json:"total"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
}

// FilesResponse represents the response for files list
type FilesResponse struct {
	Results []File `json:"results"`
	Total   int    `json:"total"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
}
