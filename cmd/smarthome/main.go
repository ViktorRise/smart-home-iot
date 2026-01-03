package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	TemperatureFile       string `json:"temperatureFile"`
	IlluminationFile      string `json:"illuminationFile"`
	LightFile             string `json:"lightFile"`
	IlluminationThreshold int    `json:"illuminationThreshold"`

	MotionFile      string `json:"motionFile"`
	SecurityLogFile string `json:"securityLogFile"`

	TelegramBotToken string `json:"telegramBotToken"`
	TelegramChatId   int64  `json:"telegramChatId"`
}

// Функция для загрузки основного конфига
func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Функция для загрузки конфигурации с токеном и chat_id
func loadLocalConfig(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var localConfig map[string]interface{}
	if err := json.Unmarshal(data, &localConfig); err != nil {
		return nil, err
	}
	return localConfig, nil
}

func readInt(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	return strconv.Atoi(s)
}

func writeInt(path string, value int) error {
	return os.WriteFile(path, []byte(strconv.Itoa(value)+"\n"), 0644)
}

func appendLog(path string, line string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line + "\n")
	return err
}

// Telegram message sender
func sendTelegram(botToken string, chatID int64, message string) error {
	if botToken == "" || chatID == 0 {
		return fmt.Errorf("telegram token/chat_id is empty")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := strings.NewReader(
		fmt.Sprintf(`{"chat_id":%d,"text":%q}`, chatID, message),
	)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram http %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func main() {
	// Загружаем основной конфиг
	cfg, err := loadConfig("config/config.json")
	if err != nil {
		fmt.Println("❌ Не удалось прочитать config/config.json:", err)
		return
	}

	// Загружаем конфиг с токеном и chat_id
	localConfig, err := loadLocalConfig("config/config.local.json")
	if err != nil {
		fmt.Println("❌ Не удалось прочитать config/config.local.json:", err)
		return
	}

	// Объединяем данные конфигов
	if token, ok := localConfig["telegramBotToken"].(string); ok {
		cfg.TelegramBotToken = token
	}
	if chatID, ok := localConfig["telegramChatId"].(float64); ok {
		cfg.TelegramChatId = int64(chatID)
	}

	// Проверка на пустые данные
	if cfg.TelegramBotToken == "" || cfg.TelegramChatId == 0 {
		fmt.Println("❌ Ошибка: token/chat_id пустые")
		return
	}

	// Чтение температуры
	temp, err := readInt(cfg.TemperatureFile)
	if err != nil {
		fmt.Println("❌ Не найден/неверный файл температуры:", cfg.TemperatureFile)
		return
	}
	fmt.Println("🌡 Температура:", temp, "°C")

	// Чтение освещенности и управление светом
	illum, err := readInt(cfg.IlluminationFile)
	if err != nil {
		fmt.Println("❌ Не найден/неверный файл освещенности:", cfg.IlluminationFile)
		return
	}

	light, err := readInt(cfg.LightFile)
	if err != nil {
		fmt.Println("❌ Не найден/неверный файл света:", cfg.LightFile)
		return
	}

	fmt.Println("💡 Освещенность:", illum)
	fmt.Println("💡 Свет (до):", light)
	fmt.Println("⚙️ Порог освещенности:", cfg.IlluminationThreshold)

	if illum < cfg.IlluminationThreshold && light == 0 {
		_ = writeInt(cfg.LightFile, 1)
		fmt.Println("🌙 Темно — включаю свет")
	} else if illum >= cfg.IlluminationThreshold && light == 1 {
		_ = writeInt(cfg.LightFile, 0)
		fmt.Println("☀️ Светло — выключаю свет")
	} else {
		fmt.Println("ℹ️ Состояние света менять не нужно")
	}

	lightAfter, _ := readInt(cfg.LightFile)
	fmt.Println("💡 Свет (после):", lightAfter)

	// Датчик движения: проверка и уведомление
	motion, err := readInt(cfg.MotionFile)
	if err != nil {
		fmt.Println("❌ Не найден/неверный файл датчика движения:", cfg.MotionFile)
		return
	}

	if motion == 1 {
		ts := time.Now().Format("2006-01-02 15:04:05")
		alert := fmt.Sprintf("%s 🚨 ALERT: обнаружено движение!", ts)

		// Оповещение в консоли
		fmt.Println("👀 Отправка уведомления в Telegram...")
		fmt.Println(alert)

		// Отправка уведомления в Telegram
		if err := sendTelegram(cfg.TelegramBotToken, cfg.TelegramChatId, alert); err != nil {
			fmt.Println("⚠️ Telegram не отправлен:", err)
		} else {
			fmt.Println("📨 Telegram отправлен")
		}

		// Логирование в файл
		if err := appendLog(cfg.SecurityLogFile, alert); err != nil {
			fmt.Println("⚠️ Не удалось записать лог:", err)
		} else {
			fmt.Println("📝 Записано в лог:", cfg.SecurityLogFile)
		}

		// Сброс датчика
		_ = writeInt(cfg.MotionFile, 0)
		fmt.Println("🔁 Датчик движения сброшен в 0")
	} else {
		fmt.Println("🛡️ Безопасность: движения нет")
	}
}
