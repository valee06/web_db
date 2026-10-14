package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type SaldoRow struct {
	ID            int
	AccountNumber string
	Period        time.Time
	Amount        string
}

type ChargeRow struct {
	ID            int
	AccountNumber string
	Period        time.Time
	ChargeDate    time.Time
	Amount        string
}

type PaymentRow struct {
	ID            int
	AccountNumber string
	Period        time.Time
	PaymentDate   time.Time
	Amount        string
}

func fetchSaldo(ctx context.Context, db *pgxpool.Pool) ([]SaldoRow, error) {
	rows, err := db.Query(ctx, `
		SELECT id, account_number, period, amount::text
		FROM saldo
		ORDER BY account_number, period DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SaldoRow
	for rows.Next() {
		var r SaldoRow
		if err := rows.Scan(&r.ID, &r.AccountNumber, &r.Period, &r.Amount); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func fetchCharges(ctx context.Context, db *pgxpool.Pool) ([]ChargeRow, error) {
	rows, err := db.Query(ctx, `
		SELECT id, account_number, period, charge_date, amount::text
		FROM charges
		ORDER BY charge_date DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ChargeRow
	for rows.Next() {
		var r ChargeRow
		if err := rows.Scan(&r.ID, &r.AccountNumber, &r.Period, &r.ChargeDate, &r.Amount); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func fetchPayments(ctx context.Context, db *pgxpool.Pool) ([]PaymentRow, error) {
	rows, err := db.Query(ctx, `
		SELECT id, account_number, period, payment_date, amount::text
		FROM payments
		ORDER BY payment_date DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PaymentRow
	for rows.Next() {
		var r PaymentRow
		if err := rows.Scan(&r.ID, &r.AccountNumber, &r.Period, &r.PaymentDate, &r.Amount); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
