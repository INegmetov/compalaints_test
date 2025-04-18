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

// структура под  ответ API
type SentimentResponse struct {
	Confidence  float64 `json:"confidence"`
	ContentType string  `json:"content_type"`
	Language    string  `json:"language"`
	Score       float64 `json:"score"`
	Sentiment   string  `json:"sentiment"`
}

var (
	apiKey          string
	sentimentURL    string
	db              *sql.DB
	defaultStatus   string
	defaultCategory string
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

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=require",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
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

	if resp.StatusCode != 200 {
		log.Printf("Ошибка от внешнего API (код %d): %s\n", resp.StatusCode, string(body))
		return "unknown"
	}

	var sentimentResp SentimentResponse
	if err := json.Unmarshal(body, &sentimentResp); err != nil {
		log.Println("Ошибка при разборе ответа:", err)
		log.Println("Тело ответа:", string(body))
		return "unknown"
	}

	log.Println("Определённая тональность:", sentimentResp.Sentiment)
	return sentimentResp.Sentiment
}

func saveComplaint(text, sentiment string) (ComplaintResponse, error) {
	timestamp := time.Now()

	query := `INSERT INTO complaints (text, status, created_at, sentiment, category)
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`

	var id int
	err := db.QueryRow(query, text, defaultStatus, timestamp, sentiment, defaultCategory).Scan(&id)
	if err != nil {
		return ComplaintResponse{}, err
	}

	return ComplaintResponse{
		ID:        id,
		Status:    defaultStatus,
		Sentiment: sentiment,
		Category:  defaultCategory,
	}, nil
}

func complaintHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req ComplaintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	sentiment := getSentiment(req.Text)
	response, err := saveComplaint(req.Text, sentiment)
	if err != nil {
		log.Println("Ошибка при вставке в БД:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/complaint", complaintHandler)
	fmt.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
