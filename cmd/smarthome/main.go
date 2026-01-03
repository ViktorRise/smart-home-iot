package main

import (
	"encoding/json"
	"fmt"
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
}

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
	// создаём папку logs/, если её нет
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

func main() {
	cfg, err := loadConfig("config/config.json")
	if err != nil {
		fmt.Println("❌ Не удалось прочитать config/config.json:", err)
		return
	}

	// Температура
	temp, err := readInt(cfg.TemperatureFile)
	if err != nil {
		fmt.Println("❌ Не найден/неверный файл температуры:", cfg.TemperatureFile)
		return
	}
	fmt.Println("🌡 Температура:", temp, "°C")

	// Освещение + свет
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

	// Безопасность: датчик движения + оповещение
	motion, err := readInt(cfg.MotionFile)
	if err != nil {
		fmt.Println("❌ Не найден/неверный файл датчика движения:", cfg.MotionFile)
		return
	}

	if motion == 1 {
		ts := time.Now().Format("2006-01-02 15:04:05")
		alert := fmt.Sprintf("%s 🚨 ALERT: обнаружено движение!", ts)

		// “Оповещение” — вывод в консоль
		fmt.Println(alert)

		// и запись в лог-файл (как реальная система)
		if err := appendLog(cfg.SecurityLogFile, alert); err != nil {
			fmt.Println("⚠️ Не удалось записать лог:", err)
		} else {
			fmt.Println("📝 Записано в лог:", cfg.SecurityLogFile)
		}

		// сброс датчика (чтобы не спамило)
		_ = writeInt(cfg.MotionFile, 0)
		fmt.Println("🔁 Датчик движения сброшен в 0")
	} else {
		fmt.Println("🛡️ Безопасность: движения нет")
	}
}
