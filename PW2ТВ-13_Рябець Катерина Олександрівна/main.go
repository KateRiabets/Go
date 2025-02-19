package main

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
)

// Структура для збереження введених  даних і результату розрахунку
type PageData struct {
	Mass     string // Маса палива
	FuelType string // Тип палива
	Result   string // Результат розрахунку
}

var tmpl *template.Template

func main() {
	var err error
	// Завантаження HTML з файлу
	tmpl, err = template.ParseFiles("template.html")
	if err != nil {
		fmt.Println("Помилка завантаження шаблону:", err)
		return
	}

	// Обробка головної сторінки
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, PageData{}) // Відображення сторінки без даних
	})

	// Обробка форми з розрахунками
	http.HandleFunc("/calculate", calculateEmissions)

	// Запуск сервера на порту 8080
	fmt.Println("Сервер запущено на http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// Функція для розрахунку викидів твердих частинок
func calculateEmissions(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()                       // Зчитування даних з форми
	massText := r.FormValue("mass")     // Отримання значення маси палива
	fuelType := r.FormValue("fuelType") // Отримання вибраного типу палива

	// Конвертація маси у число
	mass, err := strconv.ParseFloat(massText, 64)
	if err != nil || mass <= 0 {
		// Якщо введена маса некоректна,  повідомлення про помилку
		data := PageData{
			Mass:     massText,
			FuelType: fuelType,
			Result:   "Некоректна маса палива",
		}
		tmpl.Execute(w, data)
		return
	}

	// Оголошення змінних для параметрів палива
	var Q_i, A_r, G_vyn, a_vyn, eta_zu float64

	// Визначення параметрів залежно від вибраного типу палива
	switch fuelType {
	case "Донецьке газове вугілля марки ГР":
		Q_i = 20.47
		A_r = 25.20
		G_vyn = 1.5
		a_vyn = 0.8
		eta_zu = 0.985
	case "Високосірчистий мазут марки 40":
		Q_i = 40.40
		A_r = 0.15
		G_vyn = 0.0
		a_vyn = 1.0
		eta_zu = 0.985
	case "Природний газ із газопроводу Уренгой-Ужгород":
		// Якщо обраний природний газ, твердих викидів немає
		data := PageData{
			Mass:     massText,
			FuelType: fuelType,
			Result: fmt.Sprintf(`Показник емісії твердих частинок: 0 г/ГДж
Валовий викид: 0 т`),
		}
		tmpl.Execute(w, data)
		return
	}

	// Розрахунок показника емісії твердих частинок (г/ГДж)
	k_tv := (math.Pow(10, 6) / Q_i) * a_vyn * (A_r / (100 - G_vyn)) * (1 - eta_zu)

	// Розрахунок валового викиду (тонни)
	E_j := math.Pow(10, -6) * k_tv * mass * Q_i

	// Форматування результату
	result := fmt.Sprintf(`Показник емісії твердих частинок: %.2f г/ГДж
Валовий викид: %.2f т`, k_tv, E_j)

	// Передача результату у шаблон
	data := PageData{
		Mass:     massText,
		FuelType: fuelType,
		Result:   result,
	}

	tmpl.Execute(w, data)
}
