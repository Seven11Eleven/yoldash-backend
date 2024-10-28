package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/joho/godotenv"
)

type FlagResponse struct {
	Flag string `json:"flag"`
}

func generateGrade(htmlLength int) string {
	if htmlLength > 1000 {
		return fmt.Sprintf("Ваша оценка: %d/10", 9+htmlLength%2) // 9-10
	} else if htmlLength >= 100 && htmlLength <= 1000 {
		return fmt.Sprintf("Ваша оценка: %d/10", 5+htmlLength%5) // 5-9
	} else {
		return fmt.Sprintf("Ваша оценка: %d/10", 1+htmlLength%4) // 1-4
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	log.Printf("Поступил запрос на проверку сайта: %s\n", r.FormValue("url"))

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Только POST запросы", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	url := r.FormValue("url")

	if !strings.HasPrefix(url, "http") {
		http.Error(w, "Неверный формат ссылки", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Загрузка...\n")
	time.Sleep(2 * time.Second)
	fmt.Fprintf(w, "Оценка работы...\n")
	time.Sleep(2 * time.Second)

	log.Printf("Админ-бот переходит по ссылке: %s\n", url)

	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(url)

	page.MustWaitLoad()

	htmlContent := page.MustHTML()

	grade := generateGrade(len(htmlContent))

	log.Printf("Проверка сайта завершена. Оценка: %s для сайта %s\n", grade, url)

	fmt.Fprintf(w, "Проверка завершена.\n")
	fmt.Fprintf(w, "Оценка: %s\n", grade)

}

func handleFlag(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	log.Printf("Запрос на /flag от: %s\n", r.RemoteAddr)

	if !strings.HasPrefix(r.RemoteAddr, "127.0.0.1") && !strings.HasPrefix(r.RemoteAddr, "[::1]") {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		log.Println("ne proshlo")
		return
	}

	flag := os.Getenv("FLAG")
	if flag == "" {
		http.Error(w, "Флаг не настроен", http.StatusInternalServerError)
		json.NewEncoder(w).Encode(FlagResponse{Flag: "loh"})
	}

	response := FlagResponse{
		Flag: flag,
	}

	log.Println("Флаг отправлен для пользователя: ", r.RemoteAddr)

	json.NewEncoder(w).Encode(response)
}
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	log.Println("Сервер запущен на http://localhost:8080")

	http.HandleFunc("/check", handleCheck)
	http.HandleFunc("/flag", handleFlag)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
