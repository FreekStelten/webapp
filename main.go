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

type Config struct {
	Server   string `yaml:"server"`
	UserID   string `yaml:"user_id"`
	Password string `yaml:"password"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
}

type VehicleData struct {
	Name         string
	Licenceplate string
	Startdatum   string
	Einddatum    string
}

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

func lookupHandler(w http.ResponseWriter, r *http.Request) {
	// Get the licence plate value from the form submission
	licencePlate := r.FormValue("licensePlate")

	fmt.Println(licencePlate)

	config, err := loadConfig()
	if err != nil {
		log.Println(err)
		logToFile(err.Error()) // Log the error to the file
		return
	}

	// Connection information for database
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;",
		config.Server, config.UserID, config.Password, config.Port, config.Database)

	// Connect to the Azure database
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Println(err)
		logToFile(err.Error()) // Log the error to the file
		return
	}
	defer db.Close()

	// Database query for selecting all of the reservation info.
	selectLicencePlate, err := db.Query("SELECT Name, licenceplate, begindatum, Einddatum FROM slagboom_db WHERE licenceplate = @licenceplate", sql.Named("licencePlate", licencePlate))

	if err != nil {
		log.Println(err)
		logToFile(err.Error()) // Log the error to the file
		return
	}
	// Close the query, after the function has returned.
	defer selectLicencePlate.Close()
	var licencePlates []VehicleData
	var licenceplate VehicleData

	// Loop through the reservation rows and add all the reservations to a slice.
	for selectLicencePlate.Next() {
		err := selectLicencePlate.Scan(&licenceplate.Name, &licenceplate.Licenceplate, &licenceplate.Startdatum, &licenceplate.Einddatum)
		if err != nil {
			log.Println(err)
			logToFile(err.Error()) // Log the error to the file
			return
		}
		licencePlates = append(licencePlates, licenceplate)
	}
	err = selectLicencePlate.Err()
	if err != nil {
		log.Println(err)
		logToFile(err.Error()) // Log the error to the file
		return
	}

	if len(licencePlates) == 0 {
		// Kenteken niet gevonden in de database
		output := "U bent niet geregistreerd in ons park. Neem contact op met de balie voor verdere assistentie.\n"
		io.WriteString(w, output)
		return
	}

	fmt.Println(licencePlates)
	output := "Hallo " + licenceplate.Name + ", welkom op fonteyn vakantieparken. Uw kentekenplaat is: " + licenceplate.Licenceplate + ". U heeft toegang van " + licenceplate.Startdatum + " tot " + licenceplate.Einddatum + ".\n" + "U kunt nu het park oprijden.\n"
	io.WriteString(w, output)
}

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

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/lookup", lookupHandler)
	fmt.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
