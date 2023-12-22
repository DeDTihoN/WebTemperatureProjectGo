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

// функция init() вызывается автоматически при запуске программы и загружает переменные окружения из файла .env
func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {

	// Инициализация роутера Gin
	router := gin.Default()

	// Загрузка шаблонов
	router.LoadHTMLGlob("templates/*.html")

	// Обработка запроса на главную страницу
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	// Обработка запроса на получение температуры
	router.POST("/get-temperature", func(c *gin.Context) {
		city := c.PostForm("city")
		// Проверка наличия города в запросе
		if city == "" {
			c.HTML(http.StatusOK, "index.html", gin.H{"message": "Введите город"})
			return
		}

		// Получение температуры с помощью функции getTemperature
		temperature, err := getTemperature(city)
		// Проверка наличия ошибок и вывод сообщения об ошибке
		if err != nil {
			c.HTML(http.StatusOK, "index.html", gin.H{"message": err.Error()})
			return
		}
		// Вывод температуры в шаблон
		c.HTML(http.StatusOK, "index.html", gin.H{"message": fmt.Sprintf("Температура в городе %s: %s", city, temperature)})
	})

	router.Run(":8080")
}

func translateCity(russianCity string) (string, error) {
	// Получение API ключа из переменных окружения
	apiKey := os.Getenv("GOOGLE_TRANSLATE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("Google Translate API key not found")
	}

	// Инициализация клиента Google Translate API и контекста запроса к API
	ctx := context.Background()
	client, err := translate.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}

	// Вызов метода Translate API для перевода города с русского на английский язык
	resp, err := client.Translations.List([]string{russianCity}, "en").Do()
	if err != nil {
		return "", err
	}

	// Проверка наличия перевода города в ответе
	if len(resp.Translations) == 0 {
		return "", fmt.Errorf("Translation not found")
	}

	// Возвращение переведенного города
	return resp.Translations[0].TranslatedText, nil
}

func getTemperature(city string) (string, error) {
	// Получение API ключа из переменных окружения и перевод города на английский язык
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
	// закрытие тела ответа после завершения функции getTemperature
	defer resp.Body.Close()

	// Чтение тела ответа в переменную body типа []byte
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	// Преобразование тела ответа в формате JSON в map[string]interface{} и проверка наличия ошибок
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Проверка наличия ошибок в ответе и наличия данных о температуре в ответе
	if (result["cod"] != nil && result["cod"] != 200.0) || result["main"] == nil {
		return "", fmt.Errorf("Город не найден")
	}

	temperature := result["main"].(map[string]interface{})["temp"]

	// Возвращение температуры в формате string с одним знаком после запятой
	return fmt.Sprintf("%.1f°C", temperature), nil
}
