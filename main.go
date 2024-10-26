package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-adodb" // Driver to read .mdb files with ADO
)

// Measurement represents a single measurement entry from the database
type Measurement struct {
	Value     float64 `json:"value"`
	Timestamp string  `json:"timestamp"`
}

// RequestPayload holds all measurements and metadata for JSON output
type RequestPayload struct {
	Measurements []Measurement `json:"measurements"`
	Name         string        `json:"name"`
	HostDevice   int           `json:"hostDevice"`
	Device       int           `json:"device"`
	Log          float64       `json:"log"`
	Point        string        `json:"point"`
	IDEquipment  int           `json:"id_equipment"`
}

func main() {
	// Connect to the .mdb database
	db, err := sql.Open("adodb", "Provider=Microsoft.Jet.OLEDB.4.0;Data Source=Trendlog_0027400_0000000027-M-2024-03.mdb")
	if err != nil {
		log.Fatalf("Error opening .mdb file: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established successfully.")

	// Execute a query to retrieve all rows from the specified table
	query := "SELECT SampleValue, TimeOfSample FROM tblTrendlog_0027400_0000000027"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Error querying .mdb file: %v", err)
	}
	defer rows.Close()
	log.Println("Data query executed successfully.")

	// Fetch data and structure it
	var measurements []Measurement
	rowCount := 0
	for rows.Next() {
		var value float64
		var timestamp time.Time

		if err := rows.Scan(&value, &timestamp); err != nil {
			log.Printf("Error scanning row %d: %v", rowCount+1, err)
			continue
		}
		measurements = append(measurements, Measurement{
			Value:     value,
			Timestamp: timestamp.Format("2006-01-02T15:04:05Z07:00"), // Format timestamp for JSON output
		})
		log.Printf("Row %d - Retrieved value: %f, timestamp: %s\n", rowCount+1, value, timestamp.Format("2006-01-02T15:04:05Z07:00"))
		rowCount++
	}
	log.Printf("Total rows retrieved: %d", rowCount)

	// Check if measurements were retrieved
	if len(measurements) == 0 {
		log.Println("No data found in the specified table.")
		return
	}

	// Prepare the payload to send to the server
	payload := RequestPayload{
		Measurements: measurements,
		Name:         "Sample Parameter",
		HostDevice:   1001,
		Device:       501,
		Log:          1.0,
		Point:        "P1",
		IDEquipment:  42,
	}

	// Encode the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error encoding JSON: %v", err)
	}
	log.Println("JSON payload created successfully.")

	// Send the data to the server
	log.Println("Sending data to server...")
	resp, err := http.Post("https://api.sparksentry.fr/api/v1/collect/1", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error sending data: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Data sent successfully with status: %s", resp.Status)
}
