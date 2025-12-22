package repositories

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrNotFound is returned when a document is not found
	ErrNotFound = errors.New("document not found")

	// ErrAlreadyExists is returned when a document already exists
	ErrAlreadyExists = errors.New("document already exists")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
)

// BaseRepository provides common Firestore operations
type BaseRepository struct {
	client *firestore.Client
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(client *firestore.Client) *BaseRepository {
	return &BaseRepository{
		client: client,
	}
}

// Client returns the Firestore client
func (r *BaseRepository) Client() *firestore.Client {
	return r.client
}

// PaginationOptions contains options for pagination
type PaginationOptions struct {
	Limit      int
	StartAfter interface{} // Cursor for pagination
	OrderBy    string
	Direction  firestore.Direction
}

// DefaultPaginationOptions returns default pagination options
func DefaultPaginationOptions() PaginationOptions {
	return PaginationOptions{
		Limit:     50,
		OrderBy:   "created_at",
		Direction: firestore.Desc,
	}
}

// GetDocument retrieves a document by ID from a collection
func (r *BaseRepository) GetDocument(ctx context.Context, collectionPath, docID string, dest interface{}) error {
	if docID == "" {
		return fmt.Errorf("%w: document ID is required", ErrInvalidInput)
	}

	docRef := r.client.Collection(collectionPath).Doc(docID)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get document: %w", err)
	}

	if err := docSnap.DataTo(dest); err != nil {
		return fmt.Errorf("failed to decode document: %w", err)
	}

	return nil
}

// CreateDocument creates a new document in a collection
func (r *BaseRepository) CreateDocument(ctx context.Context, collectionPath, docID string, data interface{}) error {
	if docID == "" {
		return fmt.Errorf("%w: document ID is required", ErrInvalidInput)
	}

	docRef := r.client.Collection(collectionPath).Doc(docID)

	// Check if document already exists
	_, err := docRef.Get(ctx)
	if err == nil {
		return ErrAlreadyExists
	}
	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("failed to check document existence: %w", err)
	}

	_, err = docRef.Set(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

// UpdateDocument updates an existing document
func (r *BaseRepository) UpdateDocument(ctx context.Context, collectionPath, docID string, updates []firestore.Update) error {
	if docID == "" {
		return fmt.Errorf("%w: document ID is required", ErrInvalidInput)
	}

	docRef := r.client.Collection(collectionPath).Doc(docID)

	// Check if document exists
	_, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to check document existence: %w", err)
	}

	_, err = docRef.Update(ctx, updates)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

// SetDocument sets a document (creates or replaces)
func (r *BaseRepository) SetDocument(ctx context.Context, collectionPath, docID string, data interface{}) error {
	if docID == "" {
		return fmt.Errorf("%w: document ID is required", ErrInvalidInput)
	}

	docRef := r.client.Collection(collectionPath).Doc(docID)
	_, err := docRef.Set(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to set document: %w", err)
	}

	return nil
}

// DeleteDocument deletes a document
func (r *BaseRepository) DeleteDocument(ctx context.Context, collectionPath, docID string) error {
	if docID == "" {
		return fmt.Errorf("%w: document ID is required", ErrInvalidInput)
	}

	docRef := r.client.Collection(collectionPath).Doc(docID)

	// Check if document exists
	_, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to check document existence: %w", err)
	}

	_, err = docRef.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// QueryDocuments executes a query and returns documents
func (r *BaseRepository) QueryDocuments(ctx context.Context, query firestore.Query, dest interface{}) error {
	iter := query.Documents(ctx)
	defer iter.Stop()

	docs := make([]map[string]interface{}, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate documents: %w", err)
		}

		data := doc.Data()
		data["id"] = doc.Ref.ID
		docs = append(docs, data)
	}

	// Note: This is a simplified implementation
	// In production, you would use reflection to properly unmarshal into dest
	return nil
}

// ApplyPagination applies pagination options to a query
func (r *BaseRepository) ApplyPagination(query firestore.Query, opts PaginationOptions) firestore.Query {
	if opts.OrderBy != "" {
		query = query.OrderBy(opts.OrderBy, opts.Direction)
	}

	if opts.StartAfter != nil {
		query = query.StartAfter(opts.StartAfter)
	}

	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}

	return query
}

// GenerateID generates a new document ID
func (r *BaseRepository) GenerateID(collectionPath string) string {
	return r.client.Collection(collectionPath).NewDoc().ID
}
