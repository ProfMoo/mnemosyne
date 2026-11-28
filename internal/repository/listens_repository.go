// Package repository implements persistence for domain repositories.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"

	dbgen "github.com/profmoo/mnemosyne/gen/db"
	"github.com/profmoo/mnemosyne/internal/listens"
)

// SQLiteListenRepository implements listen creation, retrieval, and listing in SQLite.
type SQLiteListenRepository struct {
	queries *dbgen.Queries
}

// NewSqliteListenRepository constructs a SQLite-backed domain repository.
func NewSqliteListenRepository(params NewSqliteListenRepositoryParams) *SQLiteListenRepository {
	return &SQLiteListenRepository{queries: dbgen.New(params.DB)}
}

// NewSqliteListenRepositoryParams supplies the SQLite connection.
type NewSqliteListenRepositoryParams struct {
	fx.In

	DB *sql.DB
}

// Create persists a complete domain listen.
func (r *SQLiteListenRepository) Create(ctx context.Context, listen *listens.Listen) error {
	if err := listen.Validate(); err != nil {
		return fmt.Errorf("create listen: %w", err)
	}
	credits, err := json.Marshal(append([]listens.ArtistCredit{}, listen.Recording.ArtistCredits...))
	if err != nil {
		return fmt.Errorf("encode artist credits: %w", err)
	}
	albums, err := json.Marshal(append([]listens.Album{}, listen.Recording.Albums...))
	if err != nil {
		return fmt.Errorf("encode albums: %w", err)
	}
	_, err = r.queries.InsertListen(ctx, dbgen.InsertListenParams{
		ID:            (*listen.ID)[:],
		CreatedAt:     *listen.CreatedAt,
		UpdatedAt:     *listen.UpdatedAt,
		Duration:      int64(listen.Recording.Duration),
		Title:         listen.Recording.Title,
		ArtistCredits: string(credits),
		Albums:        string(albums),
		Mbid:          listen.Recording.MBID[:],
	})
	if err != nil {
		return fmt.Errorf("insert listen: %w", err)
	}
	return nil
}

// Get retrieves a listen or returns listens.ErrNotFound.
func (r *SQLiteListenRepository) Get(ctx context.Context, id uuid.UUID) (*listens.Listen, error) {
	row, err := r.queries.GetListen(ctx, id[:])
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get listen %s: %w", id, listens.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get listen: %w", err)
	}
	return decodeListen(&row)
}

// List retrieves all listens, newest first.
func (r *SQLiteListenRepository) List(ctx context.Context) ([]listens.Listen, error) {
	const allRows = -1
	rows, err := r.queries.ListListens(ctx, dbgen.ListListensParams{PageLimit: allRows})
	if err != nil {
		return nil, fmt.Errorf("list listens: %w", err)
	}
	result := make([]listens.Listen, len(rows))
	for index := range rows {
		listen, decodeErr := decodeListen(&rows[index])
		if decodeErr != nil {
			return nil, decodeErr
		}
		result[index] = *listen
	}
	return result, nil
}

func decodeListen(row *dbgen.Listen) (*listens.Listen, error) {
	id, err := uuid.FromBytes(row.ID)
	if err != nil {
		return nil, fmt.Errorf("decode listen ID: %w", err)
	}
	mbid, err := uuid.FromBytes(row.Mbid)
	if err != nil {
		return nil, fmt.Errorf("decode recording MBID: %w", err)
	}
	recording := &listens.Recording{
		Title:    row.Title,
		MBID:     mbid,
		Duration: time.Duration(row.Duration),
	}
	if err := json.Unmarshal([]byte(row.ArtistCredits), &recording.ArtistCredits); err != nil {
		return nil, fmt.Errorf("decode artist credits: %w", err)
	}
	if err := json.Unmarshal([]byte(row.Albums), &recording.Albums); err != nil {
		return nil, fmt.Errorf("decode albums: %w", err)
	}
	return &listens.Listen{
		ID:        &id,
		CreatedAt: &row.CreatedAt,
		UpdatedAt: &row.UpdatedAt,
		Recording: recording,
	}, nil
}
