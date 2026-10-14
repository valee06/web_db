package main

import (
	"context"
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed templates/*.html
var templatesFS embed.FS

var tmpl *template.Template

type App struct {
	db *pgxpool.Pool
}

func main() {
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=/tmp dbname=accounting_department"
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("создать пул соединений: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("подключение к БД: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	tmpl = template.Must(template.New("").Funcs(template.FuncMap{
		"fmtDate":  func(t time.Time) string { return t.Format("2006-01-02") },
		"fmtMonth": func(t time.Time) string { return t.Format("2006-01") },
	}).ParseFS(templatesFS, "templates/*.html"))

	app := &App{db: pool}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.handleIndex)
	mux.HandleFunc("/saldo", app.handleSaldo)
	mux.HandleFunc("/charges", app.handleCharges)
	mux.HandleFunc("/payments", app.handlePayments)
	mux.HandleFunc("/reports", app.handleReportsIndex)
	mux.HandleFunc("/reports/turnover", app.handleReportTurnover)
	mux.HandleFunc("/reports/apartment", app.handleReportApartment)
	mux.HandleFunc("/reports/debtors", app.handleReportDebtors)

	addr := ":8080"
	log.Printf("Server started at http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("template %s: %v", name, err)
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}
