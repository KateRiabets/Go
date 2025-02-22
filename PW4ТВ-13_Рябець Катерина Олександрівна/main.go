package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Cable struct {
	Sech    int `json:"sech"`
	UpTo3kV int `json:"up_to_3_kV"`
	SixKV   int `json:"6_kV"`
	TenKV   int `json:"10_kV"`
}

type CableData struct {
	Conductor  string  `json:"conductor"`
	Insulation string  `json:"insulation"`
	Sheath     string  `json:"sheath"`
	Cables     []Cable `json:"cables"`
}

type EconomicDensityData struct {
	Conductor    string             `json:"conductor"`
	Insulation   string             `json:"insulation"`
	Coefficients map[string]float64 `json:"coefficients"`
}

type FourFloats struct {
	First  float64
	Second float64
	Third  float64
	Fourth float64
}

type SixFloats struct {
	First  float64
	Second float64
	Third  float64
	Fourth float64
	Fifth  float64
	Sixth  float64
}

// Змінні для зберігання зчитаних з JSON даних
var allCableData []CableData
var allEconomicDensity []EconomicDensityData

// Зчитування файлу pue.json та десеріалізація
func loadCableData(filename string) ([]CableData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var cables []CableData
	err = json.Unmarshal(bytes, &cables)
	if err != nil {
		return nil, err
	}
	return cables, nil
}

// Зчитування файлу economic_density.json та десеріалізація
func loadEconomicDensityData(filename string) ([]EconomicDensityData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var densityData []EconomicDensityData
	err = json.Unmarshal(bytes, &densityData)
	if err != nil {
		return nil, err
	}
	return densityData, nil
}

// Завдання 1-2

func findSuitableCable(
	impa int,
	voltage float64,
	cablesData []CableData,
) *CableData {
	for _, cd := range cablesData {
		for _, c := range cd.Cables {
			//Вибір значення струму з полів залежно від напруги
			var current int

			switch int(voltage) {
			case 3:
				current = c.UpTo3kV
			case 6:
				current = c.SixKV
			case 10:
				current = c.TenKV
			default:
				current = 0
			}

			// Якщо кабель має ненульову пропускну здатність і вона дорівнює impa
			if current != 0 && current == impa {
				return &cd
			}
		}
	}
	return nil
}

func findEconomicCurrentDensity(
	conductor string,
	insulation string,
	timeTm float64,
	economicDensityData []EconomicDensityData,
) *float64 {
	// Перебір всіх записів
	for _, ed := range economicDensityData {
		// Перевіряємо збіг типу провідника та ізоляції (з урахуванням регістру)
		if strings.EqualFold(ed.Conductor, conductor) && strings.EqualFold(ed.Insulation, insulation) {
			// Визначаємо діапазон:
			if timeTm >= 1000.0 && timeTm <= 3000.0 {
				val := ed.Coefficients["1000_to_3000"]
				return &val
			} else if timeTm > 3000.0 && timeTm <= 5000.0 {
				val := ed.Coefficients["3000_to_5000"]
				return &val
			} else {
				val := ed.Coefficients["5000_plus"]
				return &val
			}
		}
	}
	return nil
}

func findClosestSech(
	cableData *CableData,
	thermalStability float64,
	currentVoltage float64,
) (*Cable, float64) {
	if cableData == nil {
		return nil, 0
	}
	// Пошук точного збігу
	for _, c := range cableData.Cables {
		if float64(c.Sech) == thermalStability {
			return &c, currentVoltage
		}
	}
	// Пошук найближчого більшого значення
	var closest *Cable
	for _, c := range cableData.Cables {
		if float64(c.Sech) >= thermalStability {
			if closest == nil || c.Sech < closest.Sech {
				closest = &c
			}
		}
	}
	// Якщо знайдено кабель із більшим перерізом, зменшити напругу (у прикладі: 10->6->3)
	if closest != nil {
		newVoltage := currentVoltage
		if currentVoltage == 10.0 {
			newVoltage = 6.0
		} else if currentVoltage == 6.0 {
			newVoltage = 3.0
		}
		return closest, newVoltage
	}

	return nil, 0
}

func getThermalCoefficient(insulation string) float64 {
	lower := strings.ToLower(insulation)
	switch lower {
	case "paper":
		return 92.0
	case "plastic", "rubber":
		if lower == "plastic" {
			return 75.0
		}
		if lower == "rubber" {
			return 65.0
		}
		return 75.0
	default:

		return 92.0
	}
}

func calculateResultsWithDensity(
	currentIk float64,
	timeTf float64,
	powerSm float64,
	voltage float64,
	timeTm float64,
	powerKZ float64,
	cableData []CableData,
	economicDensityData []EconomicDensityData,
) string {
	// Струм нормального режиму
	im := (powerSm / 2.0) / (math.Sqrt(3.0) * voltage)
	// Струм після аварії
	impa := int(2 * im)

	// Пошук кабелю за Iм.па
	suitableCable := findSuitableCable(impa, voltage, cableData)

	if suitableCable == nil {
		return fmt.Sprintf(
			"Струм після аварії (Iм.па): %d А\nПідходящий кабель не знайдено.",
			impa,
		)
	}
	// Економічна густина
	economicDensity := findEconomicCurrentDensity(
		suitableCable.Conductor,
		suitableCable.Insulation,
		timeTm,
		economicDensityData,
	)
	// Термічна стійкість
	thermalCoefficient := getThermalCoefficient(suitableCable.Insulation)
	thermalStability := currentIk * math.Sqrt(timeTf) / thermalCoefficient

	if economicDensity == nil {
		// Якщо не знайшли економічної густини
		return fmt.Sprintf(
			"Номінальний струм (Iном): %.2f А\n"+
				"Струм після аварії (Iм.па): %d А\n"+
				"Термічна стійкість (s): %.2f мм²\n"+
				"Підходящий кабель: провідник - %s, ізоляція - %s, оболонка - %s\n"+
				"Не вдалося знайти економічну густину струму.",
			im, impa, thermalStability,
			suitableCable.Conductor, suitableCable.Insulation, suitableCable.Sheath,
		)
	}

	// Економічний переріз
	sek := im / *economicDensity

	// Шукаємо кабель з урахуванням термічної стійкості
	closestCable, foundVoltage := findClosestSech(suitableCable, thermalStability, voltage)
	if closestCable == nil {
		return fmt.Sprintf(
			"Неможливо знайти підходящу секцію для значення термічної стійкості %.2f мм².",
			thermalStability,
		)
	}

	var uSn float64
	var ukPercent float64
	var sNomT float64
	if voltage == 10.0 {
		uSn = 10.5
		ukPercent = 10.5
	} else {
		uSn = 6.3
		ukPercent = 6.3
	}

	if foundVoltage == 10.0 {
		sNomT = 10.5
	} else {
		sNomT = 6.3
	}

	xc := (uSn * uSn) / powerKZ
	xt := (ukPercent / 100.0) * (uSn * uSn) / sNomT
	xSum := xc + xt

	// Струм КЗ
	ip0 := uSn / (math.Sqrt(3.0) * xSum)

	// Формування підсумкового тексту
	return fmt.Sprintf(
		"Номінальний струм (Iном): %.2f А\n"+
			"Струм після аварії (Iм.па): %d А\n"+
			"Термічна стійкість (s): %.2f мм²\n"+
			"Підходящий кабель: провідник - %s, ізоляція - %s, оболонка - %s\n"+
			"Економічна густина струму: %.2f A/мм²\n"+
			"Економічний переріз: %.2f мм²\n"+
			"Переріз жил кабеля: %d мм², Номінальна напруга: %.1f кВ\n"+
			"Перевірка:\n"+
			"U_с.н. = %.2f кВ\n"+
			"U_к%% = %.2f %%\n"+
			"S_ном.т = %.2f МВА\n"+
			"Xc = %.4f Ом\n"+
			"Xt = %.4f Ом\n"+
			"Сумарний опір XΣ = %.4f Ом\n"+
			"Початкове значення струму трифазного КЗ Iп0 = %.4f кА",
		im,
		impa,
		thermalStability,
		suitableCable.Conductor,
		suitableCable.Insulation,
		suitableCable.Sheath,
		*economicDensity,
		sek,
		closestCable.Sech,
		foundVoltage,
		uSn,
		ukPercent,
		sNomT,
		xc,
		xt,
		xSum,
		ip0,
	)
}

// Завдання 3

// Розрахунок реактивного опору трансформатора Xт
func calculateXtValue(uKmax float64, uVn float64, sNomT float64) float64 {
	if uKmax == 0.0 || uVn == 0.0 || sNomT == 0.0 {
		return 0.0
	}
	return (uKmax * math.Pow(uVn, 2)) / (100.0 * sNomT)
}

// Розрахунок  Xш, Zш, Xш.min та Zш.min
func calculateZshValues(rcN, xcN, rcMin, xcMin, xt float64) FourFloats {
	xSh := xcN + xt
	zSh := math.Sqrt(math.Pow(rcN, 2) + math.Pow(xSh, 2))

	xShMin := xcMin + xt
	zShMin := math.Sqrt(math.Pow(rcMin, 2) + math.Pow(xShMin, 2))

	return FourFloats{
		First:  xSh,
		Second: zSh,
		Third:  xShMin,
		Fourth: zShMin,
	}
}

// Розрахунок коефіцієнтів приведення k_р
func calculateK(uVn, uNn float64) float64 {
	if uVn == 0.0 || uNn == 0.0 {
		return 0.0
	}
	return math.Pow(uNn, 2) / math.Pow(uVn, 2)
}

// Розрахунок опорів на шинах 10 кВ та мінімальних
func calculateZshNValues(rcN float64, xSh float64, rcMin float64, xShMin float64, kPr float64) SixFloats {
	rShN := rcN * kPr
	xShN := xSh * kPr
	zShN := math.Sqrt(math.Pow(rShN, 2) + math.Pow(xShN, 2))

	rShNMin := rcMin * kPr
	xShNMin := xShMin * kPr
	zShNMin := math.Sqrt(math.Pow(rShNMin, 2) + math.Pow(xShNMin, 2))

	return SixFloats{
		First:  rShN,
		Second: xShN,
		Third:  zShN,
		Fourth: rShNMin,
		Fifth:  xShNMin,
		Sixth:  zShNMin,
	}
}

// Розрахунок  струмів трифазного та двофазного КЗ (нормальний та мінімальний режими)
func calculateI(uVn float64, zSh float64, zShMin float64) FourFloats {
	i3Sh := (uVn * 1000.0) / (math.Sqrt(3.0) * zSh)
	i2Sh := i3Sh * (math.Sqrt(3.0) / 2.0)

	i3ShMin := (uVn * 1000.0) / (math.Sqrt(3.0) * zShMin)
	i2ShMin := i3ShMin * (math.Sqrt(3.0) / 2.0)

	return FourFloats{
		First:  i3Sh,
		Second: i2Sh,
		Third:  i3ShMin,
		Fourth: i2ShMin,
	}
}

// Розрахунок  сумарних опорів  та мінімальних
func calculateZsumN(
	Rl float64, Xl float64,
	rShN float64, xShN float64,
	rShNMin float64, xShNMin float64,
) SixFloats {
	rSumN := Rl + rShN
	xSumN := Xl + xShN
	zSumN := math.Sqrt(math.Pow(rSumN, 2) + math.Pow(xSumN, 2))

	rSumNMin := Rl + rShNMin
	xSumNMin := Xl + xShNMin
	zSumNMin := math.Sqrt(math.Pow(rSumNMin, 2) + math.Pow(xSumNMin, 2))

	return SixFloats{
		First:  rSumN,
		Second: xSumN,
		Third:  zSumN,
		Fourth: rSumNMin,
		Fifth:  xSumNMin,
		Sixth:  zSumNMin,
	}
}

func calculateMain(
	uKmax float64,
	uVn float64,
	uNn float64,
	sNomT float64,
	rcN float64,
	xcN float64,
	rcMin float64,
	xcMin float64,
	r0 float64,
	x0 float64,
	section1_2 float64,
	section2_3 float64,
	section4_5 float64,
	section5_6 float64,
	section6_7 float64,
	section7_8 float64,
	section8_9 float64,
	section9_10 float64,
) string {
	// 1. Xт
	xt := calculateXtValue(uKmax, uVn, sNomT)

	// 2. Zш та Zш.min
	zshVals := calculateZshValues(rcN, xcN, rcMin, xcMin, xt)
	xSh := zshVals.First
	zSh := zshVals.Second
	xShMin := zshVals.Third
	zShMin := zshVals.Fourth

	// 3. Струми КЗ на шинах 10 кВ
	iVals := calculateI(uVn, zSh, zShMin)
	i3Sh := iVals.First
	i2Sh := iVals.Second
	i3ShMin := iVals.Third
	i2ShMin := iVals.Fourth

	// 4. Коефіцієнт приведення k_р
	kPr := calculateK(uVn, uNn)

	// 5. Опори на шинах 10 кВ (номінальний та мінімальний)
	zshNVals := calculateZshNValues(rcN, xSh, rcMin, xShMin, kPr)
	rShN := zshNVals.First
	xShN := zshNVals.Second
	zShN := zshNVals.Third
	rShNMin := zshNVals.Fourth
	xShNMin := zshNVals.Fifth
	zShNMin := zshNVals.Sixth

	// 6. Струми КЗ у нормальному та мінімальному режимі (на шинах 10 кВ, але приведені)
	iValsN := calculateI(uNn, zShN, zShNMin)
	i3ShN := iValsN.First
	i2ShN := iValsN.Second
	i3ShNMin := iValsN.Third
	i2ShNMin := iValsN.Fourth

	// 7. Сумарна довжина
	totalLength := section1_2 + section2_3 + section4_5 + section5_6 + section6_7 + section7_8 + section8_9 + section9_10

	// Rл, Xл
	Rl := totalLength * r0
	Xl := totalLength * x0

	// 8. Опори в точці 10 (ZΣ.н, мін. тощо)
	zSumVals := calculateZsumN(Rl, Xl, rShN, xShN, rShNMin, xShNMin)
	rSumN := zSumVals.First
	xSumN := zSumVals.Second
	zSumN := zSumVals.Third
	rSumNMin := zSumVals.Fourth
	xSumNMin := zSumVals.Fifth
	zSumNMin := zSumVals.Sixth

	// 9. Струми КЗ в точці 10 (номінальний/мінімальний)
	iValsL := calculateI(uNn, zSumN, zSumNMin)
	i3LN := iValsL.First
	i2LN := iValsL.Second
	i3LNMin := iValsL.Third
	i2LNMin := iValsL.Fourth

	// Формуємо текстовий звіт
	return fmt.Sprintf(`
Реактивний опір трансформатора: XТ = %.2f Ом

Xш = %.2f Ом
Zш = %.2f Ом
Xш.min = %.2f Ом
Zш.min = %.2f Ом

Струми КЗ у нормальному режимі:
I(3)ш = %.2f А
I(2)ш = %.2f А

Струми КЗ у мінімальному режимі:
I(3)ш.min = %.2f А
I(2)ш.min = %.2f А

Коефіцієнт приведення:
k_р = %.3f

Нормальний режим (приведені опори):
Rш.н = %.2f Ом
Xш.н = %.2f Ом
Zш.н = %.2f Ом

Мінімальний режим (приведені опори):
Rш.н.мін = %.2f Ом
Xш.н.мін = %.2f Ом
Zш.н.мін = %.2f Ом

Струми КЗ на шинах 10 кВ (з урахуванням приведення):
I(3)ш (норм.) = %.2f А
I(2)ш (норм.) = %.2f А
I(3)ш.min = %.2f А
I(2)ш.min = %.2f А

Сумарна довжина: %.2f км
Rл = %.2f Ом
Xл = %.2f Ом

Нормальний режим (точка 10):
RΣ.н = %.2f Ом
XΣ.н = %.2f Ом
ZΣ.н = %.2f Ом

Мінімальний режим (точка 10):
RΣ.н.мін = %.2f Ом
XΣ.н.мін = %.2f Ом
ZΣ.н.мін = %.2f Ом

Струми КЗ у точці 10 (нормальний режим):
I(3)ш = %.2f А
I(2)ш = %.2f А

Струми КЗ у точці 10 (мінімальний режим):
I(3)ш.min = %.2f А
I(2)ш.min = %.2f А
`,
		xt,
		xSh, zSh, xShMin, zShMin,
		i3Sh, i2Sh,
		i3ShMin, i2ShMin,
		kPr,
		rShN, xShN, zShN,
		rShNMin, xShNMin, zShNMin,
		i3ShN, i2ShN, i3ShNMin, i2ShNMin,
		totalLength, Rl, Xl,
		rSumN, xSumN, zSumN,
		rSumNMin, xSumNMin, zSumNMin,
		i3LN, i2LN,
		i3LNMin, i2LNMin,
	)
}

type PageData struct {
	CurrentIk string
	TimeTf    string
	PowerSm   string
	Voltage   string
	TimeTm    string
	PowerKZ   string
	Result12  string

	UKmax       string
	UVn         string
	UNn         string
	SNomT       string
	RcN         string
	XcN         string
	RcMin       string
	XcMin       string
	R0          string
	X0          string
	Section1_2  string
	Section2_3  string
	Section4_5  string
	Section5_6  string
	Section6_7  string
	Section7_8  string
	Section8_9  string
	Section9_10 string
	Result3     string
}

var tmpl *template.Template

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		// Дефолтні значення
		CurrentIk: "2500",
		TimeTf:    "2.5",
		PowerSm:   "1300",
		Voltage:   "10",
		TimeTm:    "4000",
		PowerKZ:   "2000",

		UKmax:       "11.1",
		UVn:         "115",
		UNn:         "11",
		SNomT:       "6.3",
		RcN:         "10.65",
		XcN:         "24.02",
		RcMin:       "34.88",
		XcMin:       "65.68",
		R0:          "0.64",
		X0:          "0.363",
		Section1_2:  "0.2",
		Section2_3:  "0.35",
		Section4_5:  "0.2",
		Section5_6:  "0.6",
		Section6_7:  "2.0",
		Section7_8:  "2.55",
		Section8_9:  "3.37",
		Section9_10: "3.1",
	}
	tmpl.Execute(w, data)
}

// Завдань 1-2
func handlerCalculate12(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Зчитуємо поля
		currentIk := r.FormValue("currentIk")
		timeTf := r.FormValue("timeTf")
		powerSm := r.FormValue("powerSm")
		voltage := r.FormValue("voltage")
		timeTm := r.FormValue("timeTm")
		powerKZ := r.FormValue("powerKZ")

		// Перетворюємо у float64
		fIk, _ := strconv.ParseFloat(currentIk, 64)
		fTf, _ := strconv.ParseFloat(timeTf, 64)
		fSm, _ := strconv.ParseFloat(powerSm, 64)
		fVoltage, _ := strconv.ParseFloat(voltage, 64)
		fTm, _ := strconv.ParseFloat(timeTm, 64)
		fKZ, _ := strconv.ParseFloat(powerKZ, 64)

		result := calculateResultsWithDensity(
			fIk, fTf, fSm, fVoltage, fTm, fKZ,
			allCableData, allEconomicDensity,
		)

		// Формуємо дані для шаблону:
		data := PageData{
			CurrentIk: currentIk,
			TimeTf:    timeTf,
			PowerSm:   powerSm,
			Voltage:   voltage,
			TimeTm:    timeTm,
			PowerKZ:   powerKZ,
			Result12:  result,

			// Поля Завдання 3 залишимо з дефолтами
			UKmax:       "11.1",
			UVn:         "115",
			UNn:         "11",
			SNomT:       "6.3",
			RcN:         "10.65",
			XcN:         "24.02",
			RcMin:       "34.88",
			XcMin:       "65.68",
			R0:          "0.64",
			X0:          "0.363",
			Section1_2:  "0.2",
			Section2_3:  "0.35",
			Section4_5:  "0.2",
			Section5_6:  "0.6",
			Section6_7:  "2.0",
			Section7_8:  "2.55",
			Section8_9:  "3.37",
			Section9_10: "3.1",
		}
		tmpl.Execute(w, data)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Завдання 3
func handlerCalculate3(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Зчитуємо поля
		uKmax := r.FormValue("uKmax")
		uVn := r.FormValue("uVn")
		uNn := r.FormValue("uNn")
		sNomT := r.FormValue("sNomT")
		rcN := r.FormValue("rc_n")
		xcN := r.FormValue("xc_n")
		rcMin := r.FormValue("rc_min")
		xcMin := r.FormValue("xc_min")
		r0 := r.FormValue("r0")
		x0 := r.FormValue("x0")

		section1_2 := r.FormValue("section1_2")
		section2_3 := r.FormValue("section2_3")
		section4_5 := r.FormValue("section4_5")
		section5_6 := r.FormValue("section5_6")
		section6_7 := r.FormValue("section6_7")
		section7_8 := r.FormValue("section7_8")
		section8_9 := r.FormValue("section8_9")
		section9_10 := r.FormValue("section9_10")

		// Парсимо
		fUKmax, _ := strconv.ParseFloat(uKmax, 64)
		fUVn, _ := strconv.ParseFloat(uVn, 64)
		fUNn, _ := strconv.ParseFloat(uNn, 64)
		fSNomT, _ := strconv.ParseFloat(sNomT, 64)
		fRcN, _ := strconv.ParseFloat(rcN, 64)
		fXcN, _ := strconv.ParseFloat(xcN, 64)
		fRcMin, _ := strconv.ParseFloat(rcMin, 64)
		fXcMin, _ := strconv.ParseFloat(xcMin, 64)
		fR0, _ := strconv.ParseFloat(r0, 64)
		fX0, _ := strconv.ParseFloat(x0, 64)

		fSec12, _ := strconv.ParseFloat(section1_2, 64)
		fSec23, _ := strconv.ParseFloat(section2_3, 64)
		fSec45, _ := strconv.ParseFloat(section4_5, 64)
		fSec56, _ := strconv.ParseFloat(section5_6, 64)
		fSec67, _ := strconv.ParseFloat(section6_7, 64)
		fSec78, _ := strconv.ParseFloat(section7_8, 64)
		fSec89, _ := strconv.ParseFloat(section8_9, 64)
		fSec910, _ := strconv.ParseFloat(section9_10, 64)

		result := calculateMain(
			fUKmax, fUVn, fUNn, fSNomT,
			fRcN, fXcN, fRcMin, fXcMin,
			fR0, fX0,
			fSec12, fSec23, fSec45, fSec56,
			fSec67, fSec78, fSec89, fSec910,
		)

		// Повертаємо результат у шаблон
		data := PageData{
			// Поля для Завдань 1-2 (залишимо дефолтні)
			CurrentIk: "2500",
			TimeTf:    "2.5",
			PowerSm:   "1300",
			Voltage:   "10",
			TimeTm:    "4000",
			PowerKZ:   "2000",
			Result12:  "",

			// Поля для Завдання 3
			UKmax:       uKmax,
			UVn:         uVn,
			UNn:         uNn,
			SNomT:       sNomT,
			RcN:         rcN,
			XcN:         xcN,
			RcMin:       rcMin,
			XcMin:       xcMin,
			R0:          r0,
			X0:          x0,
			Section1_2:  section1_2,
			Section2_3:  section2_3,
			Section4_5:  section4_5,
			Section5_6:  section5_6,
			Section6_7:  section6_7,
			Section7_8:  section7_8,
			Section8_9:  section8_9,
			Section9_10: section9_10,
			Result3:     result,
		}
		tmpl.Execute(w, data)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	var err error

	// Завантаження JSON-даних
	allCableData, err = loadCableData("pue.json")
	if err != nil {
		log.Fatalf("Помилка завантаження pue.json: %v", err)
	}

	allEconomicDensity, err = loadEconomicDensityData("economic_density.json")
	if err != nil {
		log.Fatalf("Помилка завантаження economic_density.json: %v", err)
	}

	// Парсимо HTML-шаблон
	tmpl, err = template.ParseFiles("template.html")
	if err != nil {
		log.Fatalf("Помилка парсингу шаблону index.html: %v", err)
	}

	// Роутінг
	http.HandleFunc("/", handlerIndex)
	http.HandleFunc("/calculate12", handlerCalculate12)
	http.HandleFunc("/calculate3", handlerCalculate3)

	fmt.Println("Сервер запущено на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
