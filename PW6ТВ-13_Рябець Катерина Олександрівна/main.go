package main

import (
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
)

type EquipmentParams struct {
	Name   string
	Eta    float64
	CosPhi float64
	UH     float64
	N      float64
	PH     float64
	KV     float64
	TgPhi  float64
}

type Results struct {
	TotalPower      float64
	WeightedPower   float64
	WeightedPowerTg float64
	Current         float64
	SquaredPower    float64
}
type PageData struct {
	EquipmentList   []EquipmentParams
	Results         map[string]Results
	GroupKv         float64
	NE              float64
	KR              float64
	Pp              float64
	Qp              float64
	Sp              float64
	Ip              float64
	GroupKvWorkshop float64
	NEWorkshop      float64
	KRWorkshop      float64
	PpWorkshop      float64
	QpWorkshop      float64
	SpWorkshop      float64
	IpWorkshop      float64
}

var equipmentList = []EquipmentParams{
	{"Шліфувальний верстат", 0.92, 0.9, 0.38, 4, 20, 0.15, 1.33},
	{"Свердлильний верстат", 0.92, 0.9, 0.38, 2, 14, 0.12, 1.00},
	{"Фігувальний верстат", 0.92, 0.9, 0.38, 4, 42, 0.15, 1.33},
	{"Циркулярна пила", 0.92, 0.9, 0.38, 1, 36, 0.3, 1.52},
	{"Прес", 0.92, 0.9, 0.38, 1, 20, 0.5, 0.75},
	{"Полірувальний верстат", 0.92, 0.9, 0.38, 1, 40, 0.2, 1.00},
	{"Фрезерний верстат", 0.92, 0.9, 0.38, 2, 32, 0.2, 1.00},
	{"Вентилятор", 0.92, 0.9, 0.38, 1, 20, 0.65, 0.75},
	{"Зварювальний трансформатор", 0.92, 0.9, 0.38, 2, 100, 0.2, 3.00},
	{"Сушильна шафа", 1.0, 1.0, 0.38, 2, 120, 0.8, 0.0},
}
var coefficientTable = [][]float64{
	{8.00, 5.33, 4.00, 2.67, 2.00, 1.60, 1.33, 1.14, 1.00},
	{6.22, 4.33, 3.06, 2.45, 1.98, 1.60, 1.33, 1.14, 1.00},
	{4.66, 2.89, 2.31, 1.74, 1.45, 1.34, 1.22, 1.14, 1.00},
	{3.24, 2.35, 1.91, 1.47, 1.25, 1.21, 1.12, 1.06, 1.00},
	{2.84, 2.09, 1.72, 1.35, 1.16, 1.16, 1.08, 1.03, 1.00},
	{2.64, 1.96, 1.62, 1.28, 1.14, 1.13, 1.06, 1.01, 1.00},
	{2.49, 1.86, 1.54, 1.23, 1.12, 1.10, 1.04, 1.00, 1.00},
	{2.37, 1.78, 1.48, 1.19, 1.10, 1.08, 1.02, 1.00, 1.00},
	{2.27, 1.71, 1.43, 1.16, 1.09, 1.07, 1.01, 1.00, 1.00},
	{2.18, 1.65, 1.39, 1.13, 1.07, 1.05, 1.00, 1.00, 1.00},
	{2.04, 1.56, 1.32, 1.08, 1.05, 1.03, 1.00, 1.00, 1.00},
	{1.94, 1.49, 1.27, 1.05, 1.02, 1.00, 1.00, 1.00, 1.00},
	{1.85, 1.43, 1.23, 1.02, 1.00, 1.00, 1.00, 1.00, 1.00},
	{1.78, 1.39, 1.19, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00},
	{1.72, 1.35, 1.16, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00},
	{1.60, 1.27, 1.10, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00},
	{1.51, 1.21, 1.05, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00},
	{1.44, 1.16, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00},
	{1.40, 1.13, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00},
	{1.30, 1.07, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00},
	{1.25, 1.03, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00},
	{1.16, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00, 1.00},
}
var rowHeaders = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 14, 16, 18, 20, 25}
var colHeaders = []float64{0.1, 0.15, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8}
var secondTable = [][]float64{
	{8.00, 5.33, 4.00, 2.67, 2.00, 1.60, 1.33, 1.14, 1.14},
	{5.01, 3.44, 2.69, 1.90, 1.52, 1.24, 1.11, 1.00, 1.00},
	{2.40, 2.17, 1.80, 1.42, 1.23, 1.14, 1.08, 1.00, 1.00},
	{2.28, 1.73, 1.46, 1.19, 1.06, 1.04, 0.97, 0.94, 0.94},
	{1.31, 1.20, 1.00, 0.96, 0.95, 0.94, 0.93, 0.91, 0.91},
	{1.10, 0.97, 0.91, 0.91, 0.90, 0.90, 0.90, 0.90, 0.90},
	{0.80, 0.80, 0.80, 0.85, 0.85, 0.85, 0.85, 0.85, 0.85},
	{0.75, 0.75, 0.75, 0.75, 0.75, 0.75, 0.85, 0.85, 0.85},
	{0.65, 0.65, 0.65, 0.70, 0.70, 0.70, 0.75, 0.80, 0.80},
}
var secondRowHeaders = []int{1, 2, 3, 4, 5, 6, 9, 10, 50}
var secondColHeaders = []float64{0.1, 0.15, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8}

func findInTable(ne int, kv float64, table [][]float64, rowHeaders []int, colHeaders []float64) float64 {
	rowIndex := len(rowHeaders) - 1 // 1. Визначаємо індекс рядка (n_e)
	for i, val := range rowHeaders {
		if ne <= val {
			rowIndex = i
			break
		}
	}
	rowPrevIndex := max(rowIndex-1, 0) // Індекс попереднього рядка для інтерполяції
	// 2. Визначаємо індекс стовпця (k_v)
	colIndex := len(colHeaders) - 1
	for j, val := range colHeaders {
		if kv <= val {
			colIndex = j
			break
		}
	}
	colPrevIndex := max(colIndex-1, 0) // Індекс попереднього стовпця для інтерполяції
	// 3. Якщо є точний збіг у таблиці, повертаємо значення без обчислень
	if ne <= rowHeaders[len(rowHeaders)-1] && rowHeaders[rowIndex] == ne &&
		kv <= colHeaders[len(colHeaders)-1] && colHeaders[colIndex] == kv {
		return table[rowIndex][colIndex]
	}
	// 4. Лінійна інтерполяція для знаходження значення між точками
	rowLower, rowUpper := rowHeaders[rowPrevIndex], rowHeaders[rowIndex] // Нижня та верхня межі по n_e
	rowFraction := 0.0
	if rowUpper != rowLower {
		rowFraction = float64(ne-rowLower) / float64(rowUpper-rowLower) // Частка для інтерполяції по n_e
	}
	colLower, colUpper := colHeaders[colPrevIndex], colHeaders[colIndex] // Нижня та верхня межі по k_v
	colFraction := 0.0
	if colUpper != colLower {
		colFraction = (kv - colLower) / (colUpper - colLower) // Частка для інтерполяції по k_v
	}
	// Отримуємо значення для чотирьох сусідніх точок
	valueLowerLower := table[rowPrevIndex][colPrevIndex] // Лівий верхній кут
	valueLowerUpper := table[rowPrevIndex][colIndex]     // Правий верхній кут
	valueUpperLower := table[rowIndex][colPrevIndex]     // Лівий нижній кут
	valueUpperUpper := table[rowIndex][colIndex]         // Правий нижній кут
	// Виконуємо двовимірну лінійну інтерполяцію
	interpolatedValue := valueLowerLower*(1-rowFraction)*(1-colFraction) +
		valueLowerUpper*(1-rowFraction)*colFraction +
		valueUpperLower*rowFraction*(1-colFraction) +
		valueUpperUpper*rowFraction*colFraction

	return interpolatedValue // Повертаємо інтерпольоване значення
}

// Розрахунки для всього цеху
func calculateWorkshopResults() (float64, float64, float64, float64, float64, float64, float64) {
	// "ВЕСЬ ЦЕХ"
	workshopTotalPower := 2330.0
	workshopWeightedPower := 752.0
	workshopWeightedPowerTg := 657.0
	workshopSquaredPower := 96399.0

	//6.1
	groupKvWorkshop := workshopWeightedPower / workshopTotalPower
	//6.2
	neWorkshop := math.Pow(workshopTotalPower, 2) / workshopSquaredPower
	//6.3
	roundedNEWorkshop := int(math.Round(neWorkshop))
	kRWorkshop := findInTable(roundedNEWorkshop, math.Round(groupKvWorkshop*10)/10, secondTable, secondRowHeaders, secondColHeaders)
	//6.4
	PpWorkshop := math.Round(kRWorkshop*10) / 10 * workshopWeightedPower
	//6.5
	QpWorkshop := math.Round(kRWorkshop*10) / 10 * workshopWeightedPowerTg
	//6.6
	SpWorkshop := math.Sqrt(PpWorkshop*PpWorkshop + QpWorkshop*QpWorkshop)
	//6.7
	IpWorkshop := PpWorkshop / 0.38

	return groupKvWorkshop, neWorkshop, kRWorkshop, PpWorkshop, QpWorkshop, SpWorkshop, IpWorkshop
}

func calculateResults() PageData {
	results := make(map[string]Results)

	// Змінні для ∑
	var sumTotalPower, sumWeightedPower, sumWeightedPowerTg, sumSquaredPower float64
	//Розрахунки для кожного ЕП
	for _, eq := range equipmentList {
		totalPower := eq.N * eq.PH                         // Розрахунок n * P_H
		weightedPower := eq.N * eq.PH * eq.KV              // Розрахунок n * P_H * k_v
		weightedPowerTg := eq.N * eq.PH * eq.KV * eq.TgPhi // Розрахунок I_p
		squaredPower := eq.N * math.Pow(eq.PH, 2)          // Розрахунок n * P_H^2
		current := totalPower / (math.Sqrt(3) * eq.UH * eq.CosPhi * eq.Eta)

		// Розрахунок сум
		if eq.Name != "Зварювальний трансформатор" && eq.Name != "Сушильна шафа" {
			sumTotalPower += totalPower           // ∑ n * P_H
			sumWeightedPower += weightedPower     // ∑ n * P_H * k_v
			sumWeightedPowerTg += weightedPowerTg // ∑ n * P_H^2
			sumSquaredPower += squaredPower       // ∑(n * P_H * k_v*tgPhi)
		}

		results[eq.Name] = Results{
			TotalPower:      totalPower,
			WeightedPower:   weightedPower,
			WeightedPowerTg: weightedPowerTg,
			Current:         current,
			SquaredPower:    squaredPower,
		}
	}
	// 4.1 Груповий коефіцієнт використання
	groupKv := sumWeightedPower / sumTotalPower

	// 4.2 Ефективна кількість ЕП:
	nE := math.Pow(sumTotalPower, 2) / sumSquaredPower

	//4.3
	roundedNE := int(math.Round(nE))
	kR := findInTable(roundedNE, math.Round(groupKv*10)/10, coefficientTable, rowHeaders, colHeaders)

	//4.4
	Pp := kR * sumWeightedPower

	//4.5
	Qp := 1.0 * sumWeightedPowerTg

	//4.6
	Sp := math.Sqrt(Pp*Pp + Qp*Qp)

	//4.7
	Ip := Pp / 0.38

	groupKv_Workshop, ne_Workshop, kR_Workshop, Pp_Workshop, Qp_Workshop, Sp_Workshop, Ip_Workshop := calculateWorkshopResults()

	return PageData{
		EquipmentList:   equipmentList,
		Results:         results,
		GroupKv:         groupKv,
		NE:              nE,
		KR:              kR,
		Pp:              Pp,
		Qp:              Qp,
		Sp:              Sp,
		Ip:              Ip,
		GroupKvWorkshop: groupKv_Workshop,
		NEWorkshop:      ne_Workshop,
		KRWorkshop:      kR_Workshop,
		PpWorkshop:      Pp_Workshop,
		QpWorkshop:      Qp_Workshop,
		SpWorkshop:      Sp_Workshop,
		IpWorkshop:      Ip_Workshop,
	}
}

func parseFloat(value string, defaultValue float64) float64 {
	if v, err := strconv.ParseFloat(value, 64); err == nil {
		return v
	}
	return defaultValue
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Завантаження шаблону
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	if r.Method == "POST" {
		for i := range equipmentList {
			equipmentList[i].Eta = parseFloat(r.FormValue("eta_"+equipmentList[i].Name), equipmentList[i].Eta)
			equipmentList[i].CosPhi = parseFloat(r.FormValue("cosphi_"+equipmentList[i].Name), equipmentList[i].CosPhi)
			equipmentList[i].UH = parseFloat(r.FormValue("uh_"+equipmentList[i].Name), equipmentList[i].UH)
			equipmentList[i].N = parseFloat(r.FormValue("n_"+equipmentList[i].Name), equipmentList[i].N)
			equipmentList[i].PH = parseFloat(r.FormValue("ph_"+equipmentList[i].Name), equipmentList[i].PH)
			equipmentList[i].KV = parseFloat(r.FormValue("kv_"+equipmentList[i].Name), equipmentList[i].KV)
			equipmentList[i].TgPhi = parseFloat(r.FormValue("tgphi_"+equipmentList[i].Name), equipmentList[i].TgPhi)
		}
	}

	// Розрахунок
	data := calculateResults()

	// Відправка HTML
	tmpl.Execute(w, data)
}

func main() {
	http.HandleFunc("/", indexHandler)
	log.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
