package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-yaml/yaml"
	_ "github.com/microsoft/go-mssqldb"
)

// Config bevat de configuratiegegevens voor de databaseverbinding
type Config struct {
	Server   string `yaml:"server"`
	UserID   string `yaml:"user_id"`
	Password string `yaml:"password"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
}

// VehicleData bevat de gegevens van een voertuig
type VehicleData struct {
	Name         string
	Licenceplate string
	Startdatum   string
	Einddatum    string
}

// LoginForm bevat het wachtwoordveld van het inlogformulier
type LoginForm struct {
	Password string
}

// loadConfig laadt de configuratie uit het YAML-bestand
func loadConfig() (*Config, error) {
	file, err := os.Open("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &config, nil
}

// connectToDatabase maakt verbinding met de database
func connectToDatabase(config *Config) (*sql.DB, error) {
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;",
		config.Server, config.UserID, config.Password, config.Port, config.Database)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// closeDatabase sluit de databaseverbinding
func closeDatabase(db *sql.DB) {
	db.Close()
}

// queryLicencePlate voert een query uit om voertuiggegevens op te halen op basis van een kentekenplaat
func queryLicencePlate(db *sql.DB, licencePlate string) ([]VehicleData, error) {
	query := "SELECT Name, licenceplate, begindatum, Einddatum FROM slagboom_db WHERE licenceplate = @licenceplate"
	rows, err := db.Query(query, sql.Named("licencePlate", licencePlate))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var licencePlates []VehicleData
	for rows.Next() {
		var licenceplate VehicleData
		err := rows.Scan(&licenceplate.Name, &licenceplate.Licenceplate, &licenceplate.Startdatum, &licenceplate.Einddatum)
		if err != nil {
			return nil, err
		}
		licencePlates = append(licencePlates, licenceplate)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return licencePlates, nil
}

// loginHandler behandelt het inlogverzoek
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "login.html")
	} else if r.Method == "POST" {
		password := r.FormValue("password")

		// Controleer of het wachtwoord correct is (pas de logica aan op basis van je vereisten)
		if password != "secret" {
			loggedIn = false
			http.Error(w, "Foutief wachtwoord opgegeven!", http.StatusUnauthorized)
			return
		} else {
			loggedIn = true
		}

		// Sla het wachtwoord op in een sessie of een cookie om de inlogstatus bij te houden

		// Redirect naar de hoofdpagina
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

var loggedIn bool

// lookupHandler behandelt het kentekenplaatzoekverzoek
func lookupHandler(w http.ResponseWriter, r *http.Request) {
	// Controleer of de gebruiker is ingelogd
	// Als niet ingelogd, omleiden naar de inlogpagina
	// Je kunt de inlogstatus opslaan in een sessie of een cookie
	// Het onderstaande voorbeeldcode gaat uit van een sessievariabele met de naam "loggedIn" om de inlogstatus te controleren
	if loggedIn == false {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Rest van de code van lookupHandler

	licencePlate := r.FormValue("licensePlate")

	config, err := loadConfig()
	if err != nil {
		log.Println(err)
		logToFile(err.Error()) // Log de fout naar het bestand
		return
	}

	db, err := connectToDatabase(config)
	if err != nil {
		log.Println(err)
		logToFile(err.Error()) // Log de fout naar het bestand
		return
	}
	defer closeDatabase(db)

	licencePlates, err := queryLicencePlate(db, licencePlate)
	if err != nil {
		log.Println(err)
		logToFile(err.Error()) // Log de fout naar het bestand
		return
	}

	if len(licencePlates) > 0 {
		licenceplate := licencePlates[0]
		output := "Hallo " + licenceplate.Name + ", welkom op fonteyn vakantieparken. Uw kentekenplaat is: " + licenceplate.Licenceplate + ". U heeft toegang van " + licenceplate.Startdatum + " tot " + licenceplate.Einddatum + ".\n" + "U kunt nu het park oprijden.\n"
		io.WriteString(w, output)
	} else {
		io.WriteString(w, "U bent niet geregistreerd in ons park. Neem contact op met de balie voor verdere assistentie.")
	}
	loggedIn = false
}

// logToFile logt de fout naar een bestand
func logToFile(msg string) {
	file, err := os.OpenFile("errors.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open error log file:", err)
		return
	}
	defer file.Close()

	log.SetOutput(file)
	log.Println(msg)
}

// serveIndexPage serveert de indexpagina
func serveIndexPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	http.HandleFunc("/", serveIndexPage)
	http.HandleFunc("/lookup", lookupHandler)
	http.HandleFunc("/login", loginHandler)
	fmt.Println("Server started on http://localhost:8080/login")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
