package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile("gpio/temperature.txt")
	if err != nil {
		fmt.Println("❌ Не найден файл gpio/temperature.txt")
		fmt.Println("Создай файл и запиши туда число, например: 22")
		return
	}

	value := strings.TrimSpace(string(data))
	fmt.Println("🌡 Температура (эмуляция датчика):", value, "°C")
}
