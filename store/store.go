package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrOutOfStock   = errors.New("หนังสือมีจำนวนไม่พอ")
	ErrInvalidOrder = errors.New("invalid order")
)

type Store struct{ Pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{Pool: pool} }

// Repository keeps HTTP handlers independent from the PostgreSQL implementation.
type Repository interface {
	ListBooks(context.Context, string, string) ([]Book, error)
	GetBook(context.Context, int64) (Book, error)
	CreateBook(context.Context, Book) (Book, error)
	UpdateBook(context.Context, Book) (Book, error)
	DeleteBook(context.Context, int64) error
	CreateMember(context.Context, string, string, string) (Member, error)
	MemberByEmail(context.Context, string) (Member, error)
	MemberByID(context.Context, int64) (Member, error)
	UpdateMemberName(context.Context, int64, string) (Member, error)
	UpdatePassword(context.Context, int64, string) error
	ListMembers(context.Context) ([]Member, error)
	CreateOrder(context.Context, int64, []OrderRequestItem) (Order, error)
	ListOrders(context.Context, *int64) ([]Order, error)
	UpdateOrderStatus(context.Context, int64, string) (Order, error)
}

type Book struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	CoverURL    string  `json:"cover_url"`
	Description string  `json:"description"`
	Stock       int32   `json:"stock"`
	Featured    bool    `json:"featured"`
}

type Member struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	PasswordHash string `json:"-"`
}

type OrderItem struct {
	BookID   int64   `json:"book_id"`
	Title    string  `json:"title"`
	Quantity int32   `json:"quantity"`
	Price    float64 `json:"price"`
}

type Order struct {
	ID        int64       `json:"id"`
	UserID    int64       `json:"user_id,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	Status    string      `json:"status"`
	Total     float64     `json:"total"`
	Items     []OrderItem `json:"items"`
}

type OrderRequestItem struct {
	BookID   int64 `json:"book_id"`
	Quantity int32 `json:"quantity"`
}
