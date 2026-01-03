package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

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

func main() {
	// 1) Температура (как было)
	temp, err := readInt("gpio/temperature.txt")
	if err != nil {
		fmt.Println("❌ Не найден/неверный gpio/temperature.txt (пример: 22)")
		return
	}
	fmt.Println("🌡 Температура (эмуляция датчика):", temp, "°C")

	// 2) Освещение (новое)
	illum, err := readInt("gpio/illumination.txt")
	if err != nil {
		fmt.Println("❌ Не найден/неверный gpio/illumination.txt (пример: 30)")
		return
	}

	light, err := readInt("gpio/light.txt")
	if err != nil {
		fmt.Println("❌ Не найден/неверный gpio/light.txt (пример: 0)")
		return
	}

	const threshold = 50 // порог: если ниже — включаем свет

	fmt.Println("💡 Освещённость:", illum)
	fmt.Println("💡 Свет (до):", light)

	if illum < threshold && light == 0 {
		_ = writeInt("gpio/light.txt", 1)
		fmt.Println("✅ Темно — включаю свет")
	} else if illum >= threshold && light == 1 {
		_ = writeInt("gpio/light.txt", 0)
		fmt.Println("✅ Светло — выключаю свет")
	} else {
		fmt.Println("ℹ️ Состояние света менять не нужно")
	}

	lightAfter, _ := readInt("gpio/light.txt")
	fmt.Println("💡 Свет (после):", lightAfter)
}
