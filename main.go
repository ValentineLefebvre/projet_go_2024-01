package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

var db *sql.DB
var store = sessions.NewCookieStore([]byte("cGPT3R0ck5_MY_Super_Secret_Key"))

type Appointment struct {
	ID        int
	Name      string
	Date      time.Time
	CreatedAt time.Time
}

type SalonOpening struct {
	ID         int
	SalonID    int
	CoiffeurID int
	DayOfWeek  int
	StartTime  time.Time
	EndTime    time.Time
}

type User struct {
	ID       int
	Username string
	Password string
	Email    string
	SalonID  int
	UserType int
}

func main() {
	var err error
	db, err = sql.Open("postgres", "postgres://user:pwd@localhost/appointmentsdb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
	DROP TABLE "appointments";
	DROP TABLE "salon";
	DROP TABLE "user";
	CREATE TABLE IF NOT EXISTS "appointments"(
		"id" SERIAL,
		"date" DATE NOT NULL,
		"salon_id" INT NOT NULL,
		"start_time" TIME(0) WITH TIME ZONE NOT NULL,
		"end_time" TIME(0) WITH TIME ZONE NOT NULL,
		"user_id" BIGINT NULL
	);
	ALTER TABLE
		"appointments" ADD PRIMARY KEY("id");
	CREATE TABLE IF NOT EXISTS "salon"(
		"id" SERIAL,
		"name" VARCHAR(255) NOT NULL,
		"adress" VARCHAR(255) NOT NULL,
		"manager_id" INT NOT NULL
	);
	ALTER TABLE
		"salon" ADD PRIMARY KEY("id");
	CREATE TABLE IF NOT EXISTS "user"(
		"id" SERIAL,
		"email" VARCHAR(255) NOT NULL,
		"pwd" VARCHAR(255) NOT NULL,
		"name" VARCHAR(255) NOT NULL,
		"admin" BOOLEAN NOT NULL
	);
	ALTER TABLE
		"user" ADD PRIMARY KEY("id");
	ALTER TABLE
		"salon" ADD CONSTRAINT "salon_manager_foreign" FOREIGN KEY("manager_id") REFERENCES "user"("id");
	ALTER TABLE
		"appointments" ADD CONSTRAINT "appointments_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "user"("id");
	ALTER TABLE
		"appointments" ADD CONSTRAINT "appointments_salon_id_foreign" FOREIGN KEY("salon_id") REFERENCES "salon"("id");

	INSERT INTO "user" (email, pwd, name, admin) VALUES ('admin@gmail.com', 'admin', 'admin', true);
	`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", showAppointments)
	http.HandleFunc("/appointments/new", showNewAppointmentForm)
	http.HandleFunc("/appointments/create", createAppointment)
	http.HandleFunc("/login", showLogin)
	http.HandleFunc("/logout", showLogout)
	http.HandleFunc("/authenticate", authenticate)
	http.HandleFunc("/signup", showSignup)
	http.HandleFunc("/create_account", createAccount)
	http.HandleFunc("/salon_openings", showSalonOpenings)
	http.HandleFunc("/salon_openings/new", showNewSalonOpeningForm)
	http.HandleFunc("/salon_openings/create", createSalonOpening)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func showAppointments(w http.ResponseWriter, r *http.Request) {
	appointments := getDBAppointments()
	renderTemplate(w, "index.html", map[string]interface{}{"appointments": appointments})
}

func showNewAppointmentForm(w http.ResponseWriter, r *http.Request) {
	// Check if the user is logged in
	isLoggedIn := isLoggedIn(r)

	// Pass the information to the template
	data := map[string]interface{}{
		"IsLoggedIn": isLoggedIn,
	}

	renderTemplate(w, "new.html", data)
}

// Function to check if the user is logged in
func isLoggedIn(r *http.Request) bool {
	session, _ := store.Get(r, "session-name")
	return session.Values["authenticated"] == true
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

func showLogin(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login.html", nil)
}

func showLogout(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "logout.html", nil)
}

func authenticate(w http.ResponseWriter, r *http.Request) {
	// Logique pour l'authentification
	// ...
	// Redirection en fonction du type d'utilisateur (salon, client, administrateur)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func showSignup(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "signup.html", nil)
}

func createAccount(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")
	salon_name := r.FormValue("salon_name")
	salon_adress := r.FormValue("salon_adress")

	var userID int
	_, err := db.Exec("INSERT INTO users (name, pwd, email, admin) VALUES ($1, $2, $3, $4) RETURNING id", username, password, email, false).Scan(&userID)
	if err != nil {
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	_, err := db.Exec("INSERT INTO salon (name, address, manager_id) VALUES ($1, $2, $3)", salon_name, salon_adress, userID)
	if err != nil {
		http.Error(w, "Failed to create salon", http.StatusInternalServerError)
		return
	}

	// http.Redirect(w, r, "/login", http.StatusSeeOther)
	http.Redirect(w, r, "/salon_openings", http.StatusSeeOther)
}

func showSalonOpenings(w http.ResponseWriter, r *http.Request) {
	salonOpenings := getDBSalonOpenings()
	renderTemplate(w, "salon_openings.html", map[string]interface{}{"salonOpenings": salonOpenings})
}

func showNewSalonOpeningForm(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "new_salon_opening.html", nil)
}

func createSalonOpening(w http.ResponseWriter, r *http.Request) {
	// Logique pour créer un créneau d'ouverture
	// ...
	http.Redirect(w, r, "/salon_openings", http.StatusSeeOther)
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

func getDBSalonOpenings() []SalonOpening {
	rows, err := db.Query("SELECT * FROM salon_openings")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var salonOpenings []SalonOpening
	for rows.Next() {
		var opening SalonOpening
		err := rows.Scan(&opening.ID, &opening.SalonID, &opening.CoiffeurID, &opening.DayOfWeek, &opening.StartTime, &opening.EndTime)
		if err != nil {
			log.Fatal(err)
		}
		salonOpenings = append(salonOpenings, opening)
	}

	return salonOpenings
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
