package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var monthFullNames = []string{
	"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
	"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь",
}

type turnoverCell struct {
	Charges  string
	Payments string
	Saldo    string
}

type turnoverAccount struct {
	AccountNumber string
	InSaldo       string
	OutSaldo      string
	Months        []turnoverCell
}

type turnoverPage struct {
	Title        string
	Year         int
	Error        string
	Rows         []turnoverAccount
	MonthHeaders []string
}

func (a *App) handleReportsIndex(w http.ResponseWriter, r *http.Request) {
	render(w, "reports", pageData{Title: "Отчёты"})
}

func (a *App) handleReportTurnover(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	year := time.Now().Year()
	if s := r.URL.Query().Get("year"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 2000 && v <= 2100 {
			year = v
		}
	}
	data := turnoverPage{
		Title:        "Оборотная ведомость",
		Year:         year,
		MonthHeaders: monthFullNames,
	}

	rows, err := a.db.Query(ctx, `
		SELECT account_number, month_num, in_saldo, charges_sum, payments_sum, out_saldo
		FROM report_turnover($1)
	`, year)
	if err != nil {
		log.Printf("report_turnover: %v", err)
		data.Error = "Ошибка запроса: " + err.Error()
		render(w, "report_turnover", data)
		return
	}
	defer rows.Close()

	order := []string{}
	byAcc := map[string]*turnoverAccount{}

	for rows.Next() {
		var acc string
		var monthNum int
		var inSaldo, charges, payments, outSaldo string
		if err := rows.Scan(&acc, &monthNum, &inSaldo, &charges, &payments, &outSaldo); err != nil {
			data.Error = "Ошибка чтения строки: " + err.Error()
			break
		}
		accRow, ok := byAcc[acc]
		if !ok {
			accRow = &turnoverAccount{
				AccountNumber: acc,
				Months:        make([]turnoverCell, 12),
			}
			byAcc[acc] = accRow
			order = append(order, acc)
		}
		if monthNum >= 1 && monthNum <= 12 {
			accRow.Months[monthNum-1] = turnoverCell{charges, payments, outSaldo}
		}
		if monthNum == 1 {
			accRow.InSaldo = inSaldo
		}
		accRow.OutSaldo = outSaldo
	}

	for _, acc := range order {
		data.Rows = append(data.Rows, *byAcc[acc])
	}

	render(w, "report_turnover", data)
}

type aptMonthRow struct {
	MonthCode string
	Charges   string
	Payments  string
}

type apartmentPage struct {
	Title         string
	Year          int
	Account       string
	Error         string
	Accounts      []string
	Rows          []aptMonthRow
	SaldoStart    string
	TotalCharges  string
	TotalPayments string
	TotalToPay    string
}

func (a *App) handleReportApartment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	year := time.Now().Year()
	if s := r.URL.Query().Get("year"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 2000 && v <= 2100 {
			year = v
		}
	}
	account := r.URL.Query().Get("account")
	data := apartmentPage{
		Title:         "Начисления и платежи",
		Year:          year,
		Account:       account,
		Rows:          []aptMonthRow{},
		SaldoStart:    "0",
		TotalCharges:  "0",
		TotalPayments: "0",
		TotalToPay:    "0",
	}

	accRows, err := a.db.Query(ctx, `
		SELECT DISTINCT account_number FROM (
			SELECT account_number FROM saldo
			UNION SELECT account_number FROM charges
			UNION SELECT account_number FROM payments
		) t ORDER BY account_number
	`)
	if err == nil {
		for accRows.Next() {
			var s string
			if accRows.Scan(&s) == nil {
				data.Accounts = append(data.Accounts, s)
			}
		}
		accRows.Close()
	}

	if account != "" {
		prevDec := fmt.Sprintf("%d-12-01", year-1)
		startOfYear := fmt.Sprintf("%d-01-01", year)
		endOfYear := fmt.Sprintf("%d-12-01", year)

		var saldoStart, totalCh, totalPay string
		err := a.db.QueryRow(ctx, `
			SELECT
				COALESCE((SELECT amount::text FROM saldo
				          WHERE account_number = $1 AND period = $2::date), '0'),
				COALESCE((SELECT SUM(amount)::text FROM charges
				          WHERE account_number = $1
				            AND period BETWEEN $3::date AND $4::date), '0'),
				COALESCE((SELECT SUM(amount)::text FROM payments
				          WHERE account_number = $1
				            AND period BETWEEN $3::date AND $4::date), '0')
		`, account, prevDec, startOfYear, endOfYear).Scan(&saldoStart, &totalCh, &totalPay)

		if err != nil {
			data.Error = "Ошибка получения итогов: " + err.Error()
		} else {
			data.SaldoStart = saldoStart
			data.TotalCharges = totalCh
			data.TotalPayments = totalPay

			fStart, _ := strconv.ParseFloat(saldoStart, 64)
			fCh, _ := strconv.ParseFloat(totalCh, 64)
			fPay, _ := strconv.ParseFloat(totalPay, 64)
			data.TotalToPay = strconv.FormatFloat(fStart+fCh-fPay, 'f', 2, 64)
		}

		rows, err := a.db.Query(ctx, `
			SELECT month_num, charges_sum, payments_sum
			FROM report_apartment($1, $2)
		`, account, year)
		if err != nil {
			data.Error = "Ошибка запроса: " + err.Error()
		} else {
			defer rows.Close()
			for rows.Next() {
				var mNum int
				var ch, pay string
				if err := rows.Scan(&mNum, &ch, &pay); err != nil {
					data.Error = err.Error()
					break
				}
				data.Rows = append(data.Rows, aptMonthRow{
					MonthCode: fmt.Sprintf("%02d", mNum),
					Charges:   ch,
					Payments:  pay,
				})
			}
		}
	}

	render(w, "report_apartment", data)
}

type debtorRow struct {
	AccountNumber string
	LastCharge    string
	Saldo         string
	Debt1         string
	Debt2         string
	Debt3         string
	DebtOver3     string
}

type debtorsPage struct {
	Title string
	AsOf  string
	Error string
	Rows  []debtorRow
}

func (a *App) handleReportDebtors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	asOf := r.URL.Query().Get("as_of")
	if asOf == "" {
		asOf = time.Now().Format("2006-01-02")
	}
	data := debtorsPage{Title: "Сводка по должникам", AsOf: asOf}

	rows, err := a.db.Query(ctx, `
		SELECT account_number, last_charge, saldo_amount,
		       debt_1, debt_2, debt_3, debt_over_3
		FROM report_debtors($1::date)
	`, asOf)
	if err != nil {
		log.Printf("report_debtors: %v", err)
		data.Error = "Ошибка запроса: " + err.Error()
		render(w, "report_debtors", data)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var row debtorRow
		if err := rows.Scan(&row.AccountNumber, &row.LastCharge, &row.Saldo,
			&row.Debt1, &row.Debt2, &row.Debt3, &row.DebtOver3); err != nil {
			data.Error = err.Error()
			break
		}
		data.Rows = append(data.Rows, row)
	}

	render(w, "report_debtors", data)
}
