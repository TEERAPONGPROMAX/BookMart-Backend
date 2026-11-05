package store

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateOrder(ctx context.Context, memberID int64, requested []OrderRequestItem) (Order, error) {
	quantities := make(map[int64]int32)
	for _, item := range requested {
		if item.BookID < 1 || item.Quantity < 1 || item.Quantity > 20 {
			return Order{}, fmt.Errorf("%w: quantity", ErrInvalidOrder)
		}
		quantities[item.BookID] += item.Quantity
		if quantities[item.BookID] > 20 {
			return Order{}, fmt.Errorf("%w: quantity", ErrInvalidOrder)
		}
	}
	ids := make([]int64, 0, len(quantities))
	for id := range quantities {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	if len(ids) == 0 {
		return Order{}, fmt.Errorf("%w: empty", ErrInvalidOrder)
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback(ctx)
	items := make([]OrderItem, 0, len(ids))
	total := 0.0
	for _, id := range ids {
		var item OrderItem
		var stock int32
		err := tx.QueryRow(ctx, `SELECT id,title,price::float8,stock FROM books WHERE id=$1 FOR UPDATE`, id).Scan(&item.BookID, &item.Title, &item.Price, &stock)
		if err != nil {
			return Order{}, mapNoRows(err)
		}
		item.Quantity = quantities[id]
		if stock < item.Quantity {
			return Order{}, fmt.Errorf("%w: %s", ErrOutOfStock, item.Title)
		}
		total += item.Price * float64(item.Quantity)
		items = append(items, item)
	}
	order := Order{UserID: memberID, Status: "pending", Total: total, Items: items}
	err = tx.QueryRow(ctx, `INSERT INTO orders(user_id,status,total) VALUES($1,'pending',$2) RETURNING id,created_at`, memberID, total).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return Order{}, err
	}
	for _, item := range items {
		_, err = tx.Exec(ctx, `INSERT INTO order_items(order_id,book_id,title,quantity,price) VALUES($1,$2,$3,$4,$5)`, order.ID, item.BookID, item.Title, item.Quantity, item.Price)
		if err != nil {
			return Order{}, err
		}
		_, err = tx.Exec(ctx, `UPDATE books SET stock=stock-$2 WHERE id=$1`, item.BookID, item.Quantity)
		if err != nil {
			return Order{}, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return Order{}, err
	}
	return order, nil
}

func (s *Store) ListOrders(ctx context.Context, memberID *int64) ([]Order, error) {
	query := `SELECT o.id,o.user_id,o.created_at,o.status,o.total::float8,oi.book_id,oi.title,oi.quantity,oi.price::float8
		FROM orders o JOIN order_items oi ON oi.order_id=o.id`
	var rows pgx.Rows
	var err error
	if memberID == nil {
		rows, err = s.Pool.Query(ctx, query+` ORDER BY o.created_at DESC,o.id,oi.id`)
	} else {
		rows, err = s.Pool.Query(ctx, query+` WHERE o.user_id=$1 ORDER BY o.created_at DESC,o.id,oi.id`, *memberID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := make([]Order, 0)
	positions := make(map[int64]int)
	for rows.Next() {
		var order Order
		var item OrderItem
		if err := rows.Scan(&order.ID, &order.UserID, &order.CreatedAt, &order.Status, &order.Total, &item.BookID, &item.Title, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		index, found := positions[order.ID]
		if !found {
			order.Items = make([]OrderItem, 0, 1)
			orders = append(orders, order)
			index = len(orders) - 1
			positions[order.ID] = index
		}
		orders[index].Items = append(orders[index].Items, item)
	}
	return orders, rows.Err()
}

func (s *Store) UpdateOrderStatus(ctx context.Context, id int64, status string) (Order, error) {
	var order Order
	err := s.Pool.QueryRow(ctx, `UPDATE orders SET status=$2 WHERE id=$1 RETURNING id,user_id,created_at,status,total::float8`, id, status).Scan(&order.ID, &order.UserID, &order.CreatedAt, &order.Status, &order.Total)
	if err != nil {
		return Order{}, mapNoRows(err)
	}
	order.Items = make([]OrderItem, 0)
	rows, err := s.Pool.Query(ctx, `SELECT book_id,title,quantity,price::float8 FROM order_items WHERE order_id=$1 ORDER BY id`, id)
	if err != nil {
		return Order{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.BookID, &item.Title, &item.Quantity, &item.Price); err != nil {
			return Order{}, err
		}
		order.Items = append(order.Items, item)
	}
	return order, rows.Err()
}
