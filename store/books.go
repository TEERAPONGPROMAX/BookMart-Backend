package store

import (
	"context"
	"strings"
)

const bookColumns = `id, title, author, category, price::float8, cover_url, description, stock, featured`

func (s *Store) ListBooks(ctx context.Context, query, category string) ([]Book, error) {
	rows, err := s.Pool.Query(ctx, `SELECT `+bookColumns+` FROM books
		WHERE ($1 = '' OR title ILIKE '%' || $1 || '%' OR author ILIKE '%' || $1 || '%')
		AND ($2 = '' OR category = $2) ORDER BY created_at DESC, id DESC`, strings.TrimSpace(query), strings.TrimSpace(category))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	books := make([]Book, 0)
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Category, &b.Price, &b.CoverURL, &b.Description, &b.Stock, &b.Featured); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, rows.Err()
}

func (s *Store) GetBook(ctx context.Context, id int64) (Book, error) {
	var b Book
	err := s.Pool.QueryRow(ctx, `SELECT `+bookColumns+` FROM books WHERE id=$1`, id).Scan(&b.ID, &b.Title, &b.Author, &b.Category, &b.Price, &b.CoverURL, &b.Description, &b.Stock, &b.Featured)
	if err != nil {
		return Book{}, mapNoRows(err)
	}
	return b, nil
}

func (s *Store) CreateBook(ctx context.Context, b Book) (Book, error) {
	err := s.Pool.QueryRow(ctx, `INSERT INTO books(title,author,category,price,cover_url,description,stock,featured)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING `+bookColumns,
		b.Title, b.Author, b.Category, b.Price, b.CoverURL, b.Description, b.Stock, b.Featured).
		Scan(&b.ID, &b.Title, &b.Author, &b.Category, &b.Price, &b.CoverURL, &b.Description, &b.Stock, &b.Featured)
	return b, err
}

func (s *Store) UpdateBook(ctx context.Context, b Book) (Book, error) {
	err := s.Pool.QueryRow(ctx, `UPDATE books SET title=$2,author=$3,category=$4,price=$5,cover_url=$6,description=$7,stock=$8,featured=$9
		WHERE id=$1 RETURNING `+bookColumns, b.ID, b.Title, b.Author, b.Category, b.Price, b.CoverURL, b.Description, b.Stock, b.Featured).
		Scan(&b.ID, &b.Title, &b.Author, &b.Category, &b.Price, &b.CoverURL, &b.Description, &b.Stock, &b.Featured)
	if err != nil {
		return Book{}, mapNoRows(err)
	}
	return b, nil
}

func (s *Store) DeleteBook(ctx context.Context, id int64) error {
	tag, err := s.Pool.Exec(ctx, `DELETE FROM books WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
