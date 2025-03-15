package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Структура даних про обладнання
type Equipment struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Omega            float64 `json:"omega"`
	TB               float64 `json:"tB"`
	Mu               float64 `json:"mu"`
	TP               float64 `json:"tP"`
	RequiresLength   bool    `json:"requiresLength"`
	RequiresQuantity bool    `json:"requiresQuantity"`
}

// Структура вхідного JSON-запиту
type RequestData struct {
	Equipment []struct {
		ID       int     `json:"id"`
		Length   float64 `json:"length,omitempty"`
		Quantity int     `json:"quantity,omitempty"`
	} `json:"equipment"`
	Lever int `json:"lever"`
}

// Структура відповіді
type ResponseData struct {
	SingleOmega       float64 `json:"singleOmega"`
	SingleTB          float64 `json:"singleTB"`
	SingleKa          float64 `json:"singleKa"`
	SingleKp          float64 `json:"singleKp"`
	DoubleOmega       float64 `json:"doubleOmega"`
	DoubleOmegaWithSW float64 `json:"doubleOmegaWithSW"`
	Conclusion        string  `json:"conclusion"`
}

// Список усього обладнання
var equipmentData = []Equipment{
	{1, "ПЛ-110 кВ", 0.007, 10.0, 0.167, 35.0, true, false},
	{2, "ПЛ-35 кВ", 0.02, 8.0, 0.167, 35.0, true, false},
	{3, "ПЛ-10 кВ", 0.02, 10.0, 0.167, 35.0, true, false},
	{4, "КЛ-10 кВ (траншея)", 0.03, 44.0, 1.0, 9.0, true, false},
	{5, "КЛ-10 кВ (кабельний канал)", 0.005, 17.5, 1.0, 9.0, true, false},
	{6, "Збірні шини 10 кВ", 0.03, 2.0, 0.167, 5.0, false, true},
	{7, "Т-110 кВ", 0.015, 100.0, 1.0, 43.0, false, false},
	{8, "Т-35 кВ", 0.02, 80.0, 1.0, 28.0, false, false},
	{9, "Т-10 кВ (кабельна мережа)", 0.005, 60.0, 0.5, 10.0, false, false},
	{10, "Т-10 кВ (повітряна мережа)", 0.05, 60.0, 0.5, 10.0, false, false},
	{11, "В-110 кВ (елегазовий)", 0.01, 30.0, 0.1, 30.0, false, false},  // Перемикач
	{12, "В-10 кВ (малооливний)", 0.02, 15.0, 0.33, 15.0, false, false}, // Перемикач
	{13, "В-10 кВ (вакуумний)", 0.01, 15.0, 0.33, 15.0, false, false},   // Перемикач
	{14, "АВ-0,38 кВ", 0.05, 4.0, 0.33, 10.0, false, false},
	{15, "ЕД 6-10 кВ", 0.1, 160.0, 0.5, 0.0, false, false},
	{16, "ЕД 0,38 кВ", 0.1, 50.0, 0.5, 0.0, false, false},
}

// Отримання об'єкта обладнання за ID
func getEquipmentByID(id int) *Equipment {
	for _, eq := range equipmentData {
		if eq.ID == id {
			return &eq
		}
	}
	return nil
}

func calculateReliability(w http.ResponseWriter, r *http.Request) {
	// Читання JSON з запиту
	requestData, err := parseJSONRequest(w, r)
	if err != nil {
		return
	}

	// Обчислення параметрів одноколової системи
	singleOmega, singleTB, singleKa, singleKp := calculateSingleCircle(requestData)

	// Обчислення параметрів двоколової системи
	doubleOmega, doubleOmegaWithSW := calculateDoubleCircle(singleOmega, singleKa, singleKp, requestData.Lever)

	// Висновок
	conclusion := determineConclusion(singleOmega, doubleOmegaWithSW)

	// Створення відповіді
	response := ResponseData{
		SingleOmega:       singleOmega,
		SingleTB:          singleTB,
		SingleKa:          singleKa,
		SingleKp:          singleKp,
		DoubleOmega:       doubleOmega,
		DoubleOmegaWithSW: doubleOmegaWithSW,
		Conclusion:        conclusion,
	}

	// Відправка JSON-відповіді
	sendJSONResponse(w, response)
}

func parseJSONRequest(w http.ResponseWriter, r *http.Request) (RequestData, error) {
	var requestData RequestData

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Помилка зчитування запиту", http.StatusBadRequest)
		return requestData, err
	}

	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(w, "Помилка декодування JSON", http.StatusBadRequest)
		return requestData, err
	}

	return requestData, nil
}

func calculateSingleCircle(requestData RequestData) (float64, float64, float64, float64) {
	var singleOmega, singleTB, singleKa, singleKp float64
	// Проходід по всіх обраних користувачем елементах обладнання
	for _, item := range requestData.Equipment {
		eq := getEquipmentByID(item.ID) // Отримання об'єкту обладнання за ID
		if eq != nil {
			length := item.Length     // Отримання довжини (якщо потрібно)
			quantity := item.Quantity // Отримання кількостві (якщо потрібно)

			// Якщо для цього обладнання не потрібно вказувати довжину,  1
			if !eq.RequiresLength {
				length = 1
			}
			// Якщо для цього обладнання не потрібно вказувати кількість,  1
			if !eq.RequiresQuantity {
				quantity = 1
			}
			// Що використовувати для множення: довжину чи кількість
			multiplier := length
			if eq.RequiresQuantity {
				multiplier = float64(quantity)
			}
			singleOmega += eq.Omega * multiplier
			singleTB += eq.TB * eq.Omega * multiplier
		}
	}
	// Якщо частота відмов не нульова, розраховуємо середній час відновлення
	if singleOmega > 0 {
		singleTB /= singleOmega
		singleKa = singleOmega * (singleTB / 8760)
		var maxTp float64 // Змінна для збереження найбільшого значення TP серед обладнання
		// Пошук максимального часу планового ремонту серед усього обладнання
		for _, item := range requestData.Equipment {
			eq := getEquipmentByID(item.ID)
			if eq != nil && eq.TP > maxTp {
				maxTp = eq.TP
			}
		}
		singleKp = 1.2 * maxTp / 8760
	}
	return singleOmega, singleTB, singleKa, singleKp
}

func calculateDoubleCircle(singleOmega, singleKa, singleKp float64, leverID int) (float64, float64) {
	doubleOmega := 2 * singleOmega * (singleKa + singleKp)

	//Отримання даних перемикача
	var switchOmega float64
	if switchEquipment := getEquipmentByID(leverID); switchEquipment != nil {
		switchOmega = switchEquipment.Omega
	}

	doubleOmegaWithSW := doubleOmega + switchOmega
	return doubleOmega, doubleOmegaWithSW
}

func determineConclusion(singleOmega, doubleOmegaWithSW float64) string {
	if singleOmega > doubleOmegaWithSW {
		return "Одноколова система менш надійна, ніж двоколова."
	}
	return "Двоколова система менш надійна, ніж одноколова."
}

func sendJSONResponse(w http.ResponseWriter, response ResponseData) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/calculate", calculateReliability)

	fmt.Println("🚀 Сервер запущено на http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
