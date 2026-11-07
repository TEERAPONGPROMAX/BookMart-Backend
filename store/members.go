package store

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateMember(ctx context.Context, email, name, passwordHash string) (Member, error) {
	var m Member
	err := s.Pool.QueryRow(ctx, `INSERT INTO members(email,display_name,password_hash) VALUES($1,$2,$3)
		RETURNING id,email,display_name,role,password_hash`, strings.ToLower(strings.TrimSpace(email)), strings.TrimSpace(name), passwordHash).
		Scan(&m.ID, &m.Email, &m.Name, &m.Role, &m.PasswordHash)
	return m, err
}

func (s *Store) MemberByEmail(ctx context.Context, email string) (Member, error) {
	var m Member
	err := s.Pool.QueryRow(ctx, `SELECT id,email,display_name,role,password_hash FROM members WHERE email=$1`, strings.ToLower(strings.TrimSpace(email))).
		Scan(&m.ID, &m.Email, &m.Name, &m.Role, &m.PasswordHash)
	if err != nil {
		return Member{}, mapNoRows(err)
	}
	return m, nil
}

func (s *Store) MemberByID(ctx context.Context, id int64) (Member, error) {
	var m Member
	err := s.Pool.QueryRow(ctx, `SELECT id,email,display_name,role,password_hash FROM members WHERE id=$1`, id).
		Scan(&m.ID, &m.Email, &m.Name, &m.Role, &m.PasswordHash)
	if err != nil {
		return Member{}, mapNoRows(err)
	}
	return m, nil
}

func (s *Store) UpdateMemberName(ctx context.Context, id int64, name string) (Member, error) {
	var m Member
	err := s.Pool.QueryRow(ctx, `UPDATE members SET display_name=$2 WHERE id=$1 RETURNING id,email,display_name,role,password_hash`, id, strings.TrimSpace(name)).
		Scan(&m.ID, &m.Email, &m.Name, &m.Role, &m.PasswordHash)
	if err != nil {
		return Member{}, mapNoRows(err)
	}
	return m, nil
}

func (s *Store) UpdatePassword(ctx context.Context, id int64, hash string) error {
	tag, err := s.Pool.Exec(ctx, `UPDATE members SET password_hash=$2 WHERE id=$1`, id, hash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListMembers(ctx context.Context) ([]Member, error) {
	rows, err := s.Pool.Query(ctx, `SELECT id,email,display_name,role FROM members ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	members := make([]Member, 0)
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Email, &m.Name, &m.Role); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (s *Store) MemberRole(ctx context.Context, id int64) (string, error) {
	var role string
	err := s.Pool.QueryRow(ctx, `SELECT role FROM members WHERE id=$1`, id).Scan(&role)
	if err != nil {
		return "", mapNoRows(err)
	}
	return role, nil
}

func (s *Store) UpdateMemberRole(ctx context.Context, id int64, role string) (Member, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Member{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `LOCK TABLE members IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return Member{}, err
	}
	var currentRole string
	if err = tx.QueryRow(ctx, `SELECT role FROM members WHERE id=$1 FOR UPDATE`, id).Scan(&currentRole); err != nil {
		return Member{}, mapNoRows(err)
	}
	if currentRole == "admin" && role == "member" {
		var adminCount int
		if err = tx.QueryRow(ctx, `SELECT count(*) FROM members WHERE role='admin'`).Scan(&adminCount); err != nil {
			return Member{}, err
		}
		if adminCount <= 1 {
			return Member{}, ErrLastAdmin
		}
	}
	var member Member
	err = tx.QueryRow(ctx, `UPDATE members SET role=$2 WHERE id=$1 RETURNING id,email,display_name,role`, id, role).
		Scan(&member.ID, &member.Email, &member.Name, &member.Role)
	if err != nil {
		return Member{}, mapNoRows(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Member{}, err
	}
	return member, nil
}

func mapNoRows(err error) error {
	if err == pgx.ErrNoRows {
		return ErrNotFound
	}
	return err
}
