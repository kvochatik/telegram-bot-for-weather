package main

import (
	"encoding/json"
	//"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"io/ioutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)


// WeatherResponse представляет структуру ответа от WeatherAPI
type WeatherResponse struct {
    Location struct {
        Name string `json:"name"`
    } `json:"location"`
    Current struct {
        TempC       float64 `json:"temp_c"`
        Condition struct {
            Text string `json:"text"`
        } `json:"condition"`
    } `json:"current"`
}

// getWeather получает погоду для указанного города
func getWeather(city string) (string, error) {
    apiKey := os.Getenv("WEATHER_API_KEY")
    if apiKey == "" {
        return "", fmt.Errorf("Апи ключ не установлен в .env файле")
    }

    // Формируем URL для запроса
    url := fmt.Sprintf("https://weatherapi-com.p.rapidapi.com/current.json?q=%s", city)

    // Создаем HTTP-запрос
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return "", fmt.Errorf("Ошибка при создание запроса: %v", err)
    }

    // Устанавливаем необходимые заголовки
    req.Header.Add("x-rapidapi-host", "weatherapi-com.p.rapidapi.com")
    req.Header.Add("x-rapidapi-key", apiKey)

    // Выполняем запрос
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("Ошибка при создание запроса: %v", err)
    }
    defer resp.Body.Close()

	//Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API вернул ошибку: %v", resp.Status)
	}

    // Читаем тело ответа
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("Ошибка при чтении ответа: %v", err)
    }

    // Декодируем JSON-ответ
    var weatherResp WeatherResponse
    if err := json.Unmarshal(body, &weatherResp); err != nil {
        return "", fmt.Errorf("Ошибка при декодировании ответа: %v", err)
    }

	//Проверяем, что данные о погоде были успешно получены
	if weatherResp.Location.Name == ""{
		return "", fmt.Errorf("Не удалось найти информацию о погоде для города: %s", city)
	}

    // Формируем ответное сообщение
    message := fmt.Sprintf("Погода в городе %s: %s, %.1f°C", weatherResp.Location.Name, weatherResp.Current.Condition.Text, weatherResp.Current.TempC)
    return message, nil
}

func main(){
	// Загружаем переменные окружения из файл .env
	err := godotenv.Load()
	if err != nil{
		log.Fatalf("Error loading .env file")
	}

	// Получаем API-токен из переменной окружения
	apiToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if apiToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set in .env file")
	}

	//Создаем новый экземпляр бота
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	//Канал для получения обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	//Обрабатываем каждое входящее сообщение
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand(){
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! я бот, которого написал Андрей. Где будем узнавать погоду?")
				bot.Send(msg)
			}
		} else {
			city := update.Message.Text
			weather, err := getWeather(city)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Не удалось получить погоду: %v", err))
				bot.Send(msg)
				continue
			}
			//отправляем сообщение с погодой
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, weather)
			bot.Send(msg)
		}
	}
}