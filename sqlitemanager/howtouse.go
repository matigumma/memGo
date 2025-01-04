package sqlitemanager

/*
import "log"

func main() {
	// Initialize SQLiteManager
	sqlManager, err := NewSQLiteManager("myhistory.db")
	if err != nil {
		log.Fatalf("Error initializing SQLiteManager: %v", err)
	}
	defer sqlManager.db.Close()

	// Add history
	err = sqlManager.AddHistory("memory123", nil, "New memory data", "ADD", nil, nil, 0)
	if err != nil {
		log.Printf("Error adding history: %v", err)
	}

	// Get history
	history, err := sqlManager.GetHistory("memory123")
	if err != nil {
		log.Printf("Error getting history: %v", err)
	}
	log.Printf("History: %+v", history)

	// Reset
	// err = sqlManager.Reset()
	// if err != nil {
	// 	log.Printf("Error resetting database: %v", err)
	// }
} */
