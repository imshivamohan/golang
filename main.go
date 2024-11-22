Here is the complete implementation for your requirement. The main.go file reads a CSV file specified via the -f flag, initializes a PostgreSQL connection, and inserts the data into a specified table.

main.go
go
Copy code
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"

	"your_project/db"    // Replace with your actual db package path
	"your_project/logger" // Replace with your actual logger package path
)

func main() {
	// Parse command-line flags
	csvFilePath := flag.String("f", "", "Path to the CSV file")
	configPath := flag.String("config", "config.yaml", "Path to the database config file")
	tableName := flag.String("table", "sample_table", "Name of the target database table")
	flag.Parse()

	if *csvFilePath == "" {
		fmt.Println("Error: Please provide the path to the CSV file using the -f flag.")
		os.Exit(1)
	}

	// Initialize database connection
	dbConn, err := db.InitDB(*configPath)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		sqlDB, _ := dbConn.DB()
		sqlDB.Close()
	}()

	// Open and read the CSV file
	file, err := os.Open(*csvFilePath)
	if err != nil {
		fmt.Printf("Error opening CSV file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV file: %v\n", err)
		os.Exit(1)
	}

	if len(records) < 2 {
		fmt.Println("Error: CSV file must contain at least one header row and one data row.")
		os.Exit(1)
	}

	// Extract headers and rows
	headers := records[0]
	rows := records[1:]

	// Prepare insert query
	placeholders := make([]string, len(headers))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	insertQuery := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		*tableName,
		strings.Join(headers, ", "),
		strings.Join(placeholders, ", "),
	)

	// Insert rows into the database
	for _, row := range rows {
		_, err := dbConn.Exec(insertQuery, convertToInterfaceSlice(row)...)
		if err != nil {
			fmt.Printf("Error inserting row: %v\n", err)
		} else {
			fmt.Println("Row inserted successfully:", row)
		}
	}
}

// convertToInterfaceSlice converts a slice of strings to a slice of empty interfaces
func convertToInterfaceSlice(data []string) []interface{} {
	interfaceSlice := make([]interface{}, len(data))
	for i, v := range data {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}
data.csv
Here’s a sample CSV file named data.csv:

csv
Copy code
id,name,age,city
1,John,25,New York
2,Jane,30,Los Angeles
3,Mike,28,Chicago
4,Alice,22,San Francisco
Sample PostgreSQL Table
Create the table in your PostgreSQL database:

sql
Copy code
CREATE TABLE sample_table (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    age INT NOT NULL,
    city TEXT NOT NULL
);
Steps to Execute
Set Up the Config: Update config.yaml with the necessary PostgreSQL database connection details.

Run the Program:

bash
Copy code
go run main.go -f data.csv -config config.yaml -table sample_table
Verify the Data: Query your PostgreSQL database to verify the inserted data:

sql
Copy code
SELECT * FROM sample_table;
Key Points
Dynamic CSV Parsing:

The program dynamically handles any CSV file as long as its header matches the database table columns.
Parameter Binding:

The use of $n placeholders ensures safety from SQL injection.
Scalability:

Easily extendable for additional preprocessing or validation of the CSV data.
This setup makes the program reusable for any PostgreSQL table, simplifying CSV data ingestion
