package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type ComplaintRequest struct {
	Text string `json:"text"`
}

type ComplaintResponse struct {
	ID        int    `json:"id"`
	Status    string `json:"status"`
	Sentiment string `json:"sentiment"`
	Category  string `json:"category"`
}

type SentimentResponse struct {
	Sentiment string `json:"sentiment"`
}

var (
	apiKey          string
	sentimentURL    string
	defaultStatus   string
	defaultCategory string

	dbUser     string
	dbPassword string
	dbHost     string
	dbPort     string
	dbName     string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	apiKey = os.Getenv("API_KEY")
	sentimentURL = os.Getenv("SENTIMENT_URL")
	defaultStatus = os.Getenv("DEFAULT_STATUS")
	defaultCategory = os.Getenv("DEFAULT_CATEGORY")

	dbUser = os.Getenv("DB_USER")
	dbPassword = os.Getenv("DB_PASSWORD")
	dbHost = os.Getenv("DB_HOST")
	dbPort = os.Getenv("DB_PORT")
	dbName = os.Getenv("DB_NAME")
}

func main() {
	http.HandleFunc("/complaint", handleComplaint)
	fmt.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleComplaint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Только POST-запросы разрешены", http.StatusMethodNotAllowed)
		return
	}

	var req ComplaintRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || strings.TrimSpace(req.Text) == "" {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	sentiment := getSentiment(req.Text)

	id, err := saveToDB(req.Text, sentiment)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	response := ComplaintResponse{
		ID:        id,
		Status:    defaultStatus,
		Sentiment: sentiment,
		Category:  defaultCategory,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getSentiment(text string) string {
	payload := map[string]string{"text": text}
	jsonBody, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", sentimentURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Println("Ошибка при создании запроса:", err)
		return "unknown"
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Ошибка при отправке запроса:", err)
		return "unknown"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var sentimentResp SentimentResponse
	if err := json.Unmarshal(body, &sentimentResp); err != nil {
		log.Println("Ошибка при разборе ответа:", err)
		return "unknown"
	}

	return sentimentResp.Sentiment
}

func saveToDB(text, sentiment string) (int, error) {
	db, err := getDBConnection()
	if err != nil {
		log.Println("Ошибка подключения к БД:", err)
		return 0, err
	}
	defer db.Close()

	var id int
	query := `
		INSERT INTO complaints (text, status, timestamp, sentiment, category)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	err = db.QueryRow(query, text, defaultStatus, time.Now(), sentiment, defaultCategory).Scan(&id)
	if err != nil {
		log.Println("Ошибка при вставке в БД:", err)
		return 0, err
	}

	return id, nil
}

func getDBConnection() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=require&pool_mode=transaction",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)
	return sql.Open("postgres", connStr)
}
