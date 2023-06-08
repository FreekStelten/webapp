package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-yaml/yaml"           //yamlfile voor config gegevens.
	_ "github.com/microsoft/go-mssqldb" //pakket wordt gebruikt om verbinding te maken met Microsoft SQL Server-databases vanuit een Go-programma.
)

// Config bevat de configuratiegegevens voor de databaseverbinding
// Config structuur die de configuratiegegevens bevat voor de databaseverbinding. De velden in de structuur worden
// geannoteerd met yaml:"..." om de bijbehorende YAML-sleutels aan te geven.
type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Login    LoginConfig    `yaml:"login"`
}

type DatabaseConfig struct {
	Server   string `yaml:"server"`
	UserID   string `yaml:"user_id"`
	Password string `yaml:"password"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
}

type LoginConfig struct {
	Hash string `yaml:"hash"`
	Salt string `yaml:"salt"`
}

// VehicleData bevat de gegevens van een voertuig
// VehicleData structuur die de gegevens van een voertuig bevat, zoals naam, kentekenplaat, startdatum en einddatum.
type VehicleData struct {
	Name         string
	Licenceplate string
	Startdatum   string
	Einddatum    string
}

// Dit is de LoginForm structuur die het wachtwoordveld van het inlogformulier bevat.
type LoginForm struct {
	Password string
}

// De loadConfig functie opent "config.yaml" bestand en decodeert de inhoud naar een Config structuur.
// bij een fout optreedt bij het openen of decoderen van het bestand, wordt een fout geretourneerd.
// Als er geen fouten optreden, wordt een verwijzing naar de Config structuur geretourneerd.
func loadConfig() (*Config, error) {
	file, err := os.Open("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(&ConfigYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &ConfigYAML, nil
}

// connectToDatabase maakt verbinding met de database
// De connectToDatabase functie maakt verbinding met azure database met behulp van de opgegeven Config structuur. Het genereert een
// verbindingsreeks op basis van de configuratiegegevens en opent een databaseverbinding met behulp van de "sqlserver" driver.
// Als er een fout optreedt bij het verbinden met de database, wordt een fout geretourneerd. Anders wordt een verwijzing naar de sql.DB structuur geretourneerd.
func connectToDatabase() (*sql.DB, error) {
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;",
		ConfigYAML.Database.Server, ConfigYAML.Database.UserID, ConfigYAML.Database.Password, ConfigYAML.Database.Port, ConfigYAML.Database.Database)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// closeDatabase sluit hier de azure-databaseverbinding
func closeDatabase(db *sql.DB) {
	db.Close()
}

// De queryLicencePlate functie voert een query uit op de database om voertuiggegevens op te halen op basis van een kentekenplaat.
// Het bereidt de query voor met een parameter @licenceplate en voert vervolgens de query uit met de opgegeven kentekenplaatwaarde.
// Het scant de resultaten van de query in een slice van VehicleData en retourneert deze. Als er een fout optreedt tijdens het uitvoeren
// van de query of het scannen van de resultaten, wordt een fout geretourneerd.
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


// De loginHandler-functie verwerkt het inlogproces. Bij een "GET"-verzoek wordt de inlogpagina weergegeven. Bij een "POST"-verzoek wordt het ingediende wachtwoord gehasht 
// en vergeleken met de opgeslagen hash-waarde. Als ze overeenkomen, wordt de gebruiker ingelogd en anders krijgt hij een foutmelding.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "login.html")
	} else if r.Method == "POST" {
		password := r.FormValue("password")
		passwordHash := generatePasswordHash(password)

		if passwordHash != ConfigYAML.Login.Hash {
			loggedIn = false
			http.Error(w, "Foutief wachtwoord opgegeven!", http.StatusUnauthorized)
			return
		}

		loggedIn = true
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// De functie generatePasswordHash neemt een wachtwoord als invoer en retourneert de gehashte versie ervan. Het wachtwoord wordt gecombineerd met een geheim zout 
// en vervolgens wordt de SHA-256 hashfunctie toegepast op de resulterende byte-array. Het resultaat wordt teruggegeven als een hexadecimale string.
func generatePasswordHash(password string) string {
	passwordByte := []byte(password + ConfigYAML.Login.Salt)
	passwordHashByte := sha256.Sum256(passwordByte)
	return fmt.Sprintf("%x", passwordHashByte[:])
}
// Dit is een variabele die de inlogstatus bijhoudt.
var loggedIn bool

// lookupHandler behandelt het kentekenplaatzoekverzoek
func lookupHandler(w http.ResponseWriter, r *http.Request) {
	//hier wordt gecontroleerd of de gebruiker is ingelogd. Als dat niet het geval is, wordt de gebruiker doorgestuurd
	//naar de login-pagina. Vervolgens wordt de waarde van de licensePlate-parameter uit het verzoek gehaald.
	if loggedIn == false {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	licencePlate := r.FormValue("licensePlate")

	// hier wordt de functie connectToDatabase aangeroepen om verbinding te maken met de azure database met behulp van de geladen
	// configuratiegegevens. Als er een fout optreedt, wordt de fout gelogd en de functie gestopt. De databaseverbinding wordt ook
	// uitgesteld gesloten met behulp van defer om ervoor te zorgen dat de verbinding uiteindelijk wordt gesloten.
	db, err := connectToDatabase()
	if err != nil {
		log.Println(err)
		logToFile(err.Error()) // Log de fout naar het bestand
		return
	}
	defer closeDatabase(db)

	// Hier wordt de functie queryLicencePlate aangeroepen om de licentieplaten op te zoeken in de database met behulp van de verkregen
	// databaseverbinding en de opgegeven kentekenplaat. Als er een fout optreedt, wordt de fout gelogd en de functie gestopt.
	licencePlates, err := queryLicencePlate(db, licencePlate)
	if err != nil {
		log.Println(err)
		logToFile(err.Error()) // Log de fout naar het bestand
		return
	}

	// Als er kenteken worden gevonden in de database, wordt er een Welkomsbericht samengesteld met de gegevens van de
	// eerste kentekenplaat en naar de HTTP-respons geschreven. Anders wordt er een andere boodschap geschreven om aan te geven dat
	// de gebruiker niet geregistreerd is in het park. Vervolgens wordt de variabele loggedIn ingesteld op false. zodat ze opnieuw moeten inloggen voor het opnieuw invoeren.
	if len(licencePlates) > 0 {
		licenceplate := licencePlates[0]
		output := "Hallo " + licenceplate.Name + ", welkom op fonteyn vakantieparken. Uw kentekenplaat is: " + licenceplate.Licenceplate + ". U heeft toegang van " + licenceplate.Startdatum + " tot " + licenceplate.Einddatum + ".\n" + "U kunt nu het park oprijden.\n"
		io.WriteString(w, output)
	} else {
		io.WriteString(w, "U bent niet geregistreerd in ons park. Neem contact op met de balie voor verdere assistentie.")
	}
	loggedIn = false
}

// De functie logToFile wordt gebruikt om een foutbericht naar een bestand te loggen. Het opent een bestand genaamd "errors.txt"
// (of maakt het bestand als het nog niet bestaat) en schrijft de error naar het bestand.
// Als er een fout optreedt bij het openen van het bestand, wordt die error in terminal getoond.
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

// De functie serveIndexPage wordt gebruikt om de indexpagina te serveren. Het reageert op het verzoek door het bestand "index.html" naar de HTTP-respons te schrijven.
func serveIndexPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

var ConfigYAML Config

func LoadYaml() {
	// Hier wordt de functie loadConfig aangeroepen om de configuratiegegevens te laden.
	// Als er een fout optreedt, wordt de fout gelogd en de functie gestopt.
	_, err := loadConfig()
	if err != nil {
		log.Println(err)
		logToFile(err.Error()) // Log de fout naar het bestand
		return
	}
}

// De main-functie is het startpunt van het programma. Het configureert de verschillende HTTP-handlers voor de verschillende routes ("/", "/lookup" en "/login").
// Het drukt ook een bericht af om aan te geven dat de server is gestart op "http://localhost:8080/login". Ten slotte start het de HTTP-server met behulp van
// http.ListenAndServe en logt eventuele fouten die optreden tijdens het uitvoeren van de server.
func main() {
	loadConfig()
	http.HandleFunc("/", serveIndexPage)
	http.HandleFunc("/lookup", lookupHandler)
	http.HandleFunc("/login", loginHandler)
	fmt.Println("Server started on http://localhost:80/login")
	log.Fatal(http.ListenAndServe(":80", nil))
}
