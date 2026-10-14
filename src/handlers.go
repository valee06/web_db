package main

import (
	"log"
	"net/http"
	"time"
)

type pageData struct {
	Title   string
	Error   string
	Success string
	Rows    any
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	render(w, "index", pageData{Title: "Главная"})
}

func (a *App) handleSaldo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := pageData{Title: "Сальдо"}

	if r.Method == http.MethodPost {
		acc := r.FormValue("account_number")
		period := r.FormValue("period")
		amount := r.FormValue("amount")

		_, err := a.db.Exec(ctx, `
			INSERT INTO saldo (account_number, period, amount)
			VALUES ($1, $2, $3::numeric)
			ON CONFLICT (account_number, period)
			DO UPDATE SET amount = EXCLUDED.amount
		`, acc, period+"-01", amount)

		if err != nil {
			data.Error = "Ошибка: " + err.Error()
		} else {
			data.Success = "Запись сохранена"
		}
	}

	rows, err := fetchSaldo(ctx, a.db)
	if err != nil {
		log.Printf("fetchSaldo: %v", err)
		data.Error = "Не удалось загрузить данные"
	}
	data.Rows = rows

	render(w, "saldo", data)
}

func (a *App) handleCharges(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := pageData{Title: "Начисления"}

	if r.Method == http.MethodPost {
		acc := r.FormValue("account_number")
		period := r.FormValue("period")
		chargeDate := r.FormValue("charge_date")
		amount := r.FormValue("amount")

		if chargeDate == "" {
			chargeDate = time.Now().Format("2006-01-02")
		}

		_, err := a.db.Exec(ctx, `
			INSERT INTO charges (account_number, period, charge_date, amount)
			VALUES ($1, $2, $3, $4::numeric)
		`, acc, period+"-01", chargeDate, amount)

		if err != nil {
			data.Error = "Ошибка: " + err.Error()
		} else {
			data.Success = "Начисление добавлено"
		}
	}

	rows, err := fetchCharges(ctx, a.db)
	if err != nil {
		log.Printf("fetchCharges: %v", err)
		data.Error = "Не удалось загрузить данные"
	}
	data.Rows = rows

	render(w, "charges", data)
}

func (a *App) handlePayments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := pageData{Title: "Платежи"}

	if r.Method == http.MethodPost {
		acc := r.FormValue("account_number")
		period := r.FormValue("period")
		paymentDate := r.FormValue("payment_date")
		amount := r.FormValue("amount")

		if paymentDate == "" {
			paymentDate = time.Now().Format("2006-01-02")
		}

		_, err := a.db.Exec(ctx, `
			INSERT INTO payments (account_number, period, payment_date, amount)
			VALUES ($1, $2, $3, $4::numeric)
		`, acc, period+"-01", paymentDate, amount)

		if err != nil {
			data.Error = "Ошибка: " + err.Error()
		} else {
			data.Success = "Платёж добавлен"
		}
	}

	rows, err := fetchPayments(ctx, a.db)
	if err != nil {
		log.Printf("fetchPayments: %v", err)
		data.Error = "Не удалось загрузить данные"
	}
	data.Rows = rows

	render(w, "payments", data)
}
