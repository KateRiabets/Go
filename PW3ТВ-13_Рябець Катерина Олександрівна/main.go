package main

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
)

// Структура для збереження введених даних і результатів розрахунку
type PageData struct {
	DailyPower    string // Середньодобова потужність Pc
	CurrentStdDev string // Поточне sigma1
	FutureStdDev  string // Майбутнє sigma2
	EnergyCost    string // Вартість електроенергії V
	ResultBefore  string // До вдосконалення
	ResultAfter   string // Після вдосконалення
	ErrorMessage  string // Помилка
}

var tmpl *template.Template

func main() {
	var err error
	// Завантаження HTML-шаблону
	tmpl, err = template.ParseFiles("template.html")
	if err != nil {
		fmt.Println("Помилка завантаження шаблону:", err)
		return
	}

	// Головна сторінка
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, PageData{})
	})

	// Обробка форми з розрахунками
	http.HandleFunc("/calculate", calculateEnergy)

	// Запуск сервера
	fmt.Println("Сервер запущено на http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// Функція для розрахунку енергії
func calculateEnergy(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// Отримання значень з форми
	dailyPower := r.FormValue("dailyPower")
	currentStdDev := r.FormValue("currentStdDev")
	futureStdDev := r.FormValue("futureStdDev")
	energyCost := r.FormValue("energyCost")

	// Перевірка, чи всі поля заповнені
	Pc, err1 := strconv.ParseFloat(dailyPower, 64)
	sigma1, err2 := strconv.ParseFloat(currentStdDev, 64)
	sigma2, err3 := strconv.ParseFloat(futureStdDev, 64)
	V, err4 := strconv.ParseFloat(energyCost, 64)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		tmpl.Execute(w, PageData{
			DailyPower:    dailyPower,
			CurrentStdDev: currentStdDev,
			FutureStdDev:  futureStdDev,
			EnergyCost:    energyCost,
			ErrorMessage:  "Будь ласка, введіть правильні числові значення!",
		})
		return
	}

	P_lower := Pc - sigma2 // Нижня межа
	P_upper := Pc + sigma2 // Верхня межа

	// Розрахунки до вдосконалення
	deltaW1 := integrateNormalDistribution(Pc, sigma1, P_lower, P_upper) // Інтегрування
	W1 := Pc * 24 * deltaW1                                              // Енергія без небалансів
	profitBefore := W1 * V                                               // Прибуток від  енергії
	W2 := Pc * 24 * (1 - deltaW1)                                        // Енергія з небалансами
	penaltyBefore := W2 * V                                              // Штраф за небаланси
	finalProfitBefore := profitBefore - penaltyBefore                    // Загальний прибуток до вдосконалення

	// Розрахунки після вдосконалення
	deltaW2 := integrateNormalDistribution(Pc, sigma2, P_lower, P_upper) // Інтегрування
	W3 := Pc * 24 * deltaW2                                              // Енергія без небалансів
	profitAfter := W3 * V                                                // Прибуток від  енергії
	W4 := Pc * 24 * (1 - deltaW2)                                        // Енергія з небалансами
	penaltyAfter := W4 * V                                               // Штраф за небаланси
	finalProfitAfter := profitAfter - penaltyAfter                       // Загальний прибуток після вдосконалення

	// Формування результату
	resultBefore := fmt.Sprintf(`До вдосконалення системи:
Частка енергії без небалансів: %.2f МВт·год
Прибуток: %.2f тис. грн
Штраф: %.2f тис. грн
Загальний прибуток: %.2f тис. грн`, W1, profitBefore, penaltyBefore, finalProfitBefore)

	resultAfter := fmt.Sprintf(`Після вдосконалення системи:
Частка енергії без небалансів: %.2f МВт·год
Прибуток: %.2f тис. грн
Штраф: %.2f тис. грн
Загальний прибуток: %.2f тис. грн`, W3, profitAfter, penaltyAfter, finalProfitAfter)

	// Передача даних у шаблон
	tmpl.Execute(w, PageData{
		DailyPower:    dailyPower,
		CurrentStdDev: currentStdDev,
		FutureStdDev:  futureStdDev,
		EnergyCost:    energyCost,
		ResultBefore:  resultBefore,
		ResultAfter:   resultAfter,
	})
}

// Функція чисельного інтегрування нормального розподілу
func integrateNormalDistribution(Pc, stdDev, P_lower, P_upper float64) float64 {
	n := 1000 // Кількість кроків для інтегрування
	step := (P_upper - P_lower) / float64(n)
	area := 0.0

	for i := 0; i < n; i++ {
		x1 := P_lower + float64(i)*step
		x2 := P_lower + float64(i+1)*step
		y1 := normalDistribution(x1, Pc, stdDev)
		y2 := normalDistribution(x2, Pc, stdDev)
		area += 0.5 * (y1 + y2) * step // Метод трапецій
	}

	return area
}

// Функція нормального розподілу
func normalDistribution(p, Pc, stdDev float64) float64 {
	return (1 / (stdDev * math.Sqrt(2*math.Pi))) * math.Exp(-math.Pow(p-Pc, 2)/(2*math.Pow(stdDev, 2)))
}
