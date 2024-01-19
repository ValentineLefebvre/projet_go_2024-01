package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

type Appointment struct {
	ID        int
	Name      string
	Date      time.Time
	CreatedAt time.Time
}

func main() {
	var err error
	db, err = sql.Open("postgres", "postgres://achraf:ok@localhost/appointmentsdb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS appointments (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			date TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", showAppointments)
	http.HandleFunc("/appointments/new", showNewAppointmentForm)
	http.HandleFunc("/appointments/create", createAppointment)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func showAppointments(w http.ResponseWriter, r *http.Request) {
	appointments := getDBAppointments()
	renderTemplate(w, "index.html", map[string]interface{}{"appointments": appointments})
}

func showNewAppointmentForm(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "new.html", nil)
}

func createAppointment(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	dateString := r.FormValue("date")

	date, err := time.Parse("2006-01-02T15:04", dateString)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO appointments (name, date) VALUES ($1, $2)", name, date)
	if err != nil {
		http.Error(w, "Failed to create appointment", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getDBAppointments() []Appointment {
	rows, err := db.Query("SELECT * FROM appointments ORDER BY date ASC")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var appointments []Appointment
	for rows.Next() {
		var app Appointment
		err := rows.Scan(&app.ID, &app.Name, &app.Date, &app.CreatedAt)
		if err != nil {
			log.Fatal(err)
		}
		appointments = append(appointments, app)
	}

	return appointments
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmplFile := fmt.Sprintf("templates/%s", tmpl)
	t, err := template.ParseFiles(tmplFile)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err)
	}
}

