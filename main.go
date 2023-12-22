package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/translate/v2"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {

	router := gin.Default()

	router.LoadHTMLGlob("templates/*.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	router.POST("/get-temperature", func(c *gin.Context) {
		city := c.PostForm("city")
		if city == "" {
			c.HTML(http.StatusOK, "index.html", gin.H{"message": "Введите город"})
			return
		}

		temperature, err := getTemperature(city)
		if err != nil {
			c.HTML(http.StatusOK, "index.html", gin.H{"message": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{"message": fmt.Sprintf("Температура в городе %s: %s", city, temperature)})
	})

	router.Run(":8080")
}

func translateCity(russianCity string) (string, error) {

	apiKey := os.Getenv("GOOGLE_TRANSLATE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("Google Translate API key not found")
	}

	// Инициализация клиента Google Translate API
	ctx := context.Background()
	client, err := translate.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}

	// Вызов метода Translate
	resp, err := client.Translations.List([]string{russianCity}, "en").Do()
	if err != nil {
		return "", err
	}

	// Проверка наличия перевода
	if len(resp.Translations) == 0 {
		return "", fmt.Errorf("Translation not found")
	}

	// Возвращение переведенного города
	return resp.Translations[0].TranslatedText, nil
}

func getTemperature(city string) (string, error) {
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	city, err := translateCity(city)
	if err != nil {
		return "", err
	}
	if apiKey == "" {
		return "", fmt.Errorf("OpenWeatherMap API key not found")
	}

	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", city, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if (result["cod"] != nil && result["cod"] != 200.0) || result["main"] == nil {
		return "", fmt.Errorf("Город не найден")
	}

	temperature := result["main"].(map[string]interface{})["temp"]

	return fmt.Sprintf("%.1f°C", temperature), nil
}
