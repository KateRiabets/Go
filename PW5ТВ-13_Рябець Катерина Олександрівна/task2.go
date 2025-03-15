package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// Структура для передачи даних в HTML-шаблон
type PageData struct {
	Omega  string
	TB     string
	PM     string
	KP     string
	TM     string
	ZPerA  string
	ZPerP  string
	Result string
}

var tmpl *template.Template

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		// Значення за замовчуванням
		Omega: "0.1",
		TB:    "0.045",
		PM:    "5120",
		KP:    "0.004",
		TM:    "6451",
		ZPerA: "23.6",
		ZPerP: "17.6",
	}
	tmpl.Execute(w, data)
}

// Розрахунки
func handlerCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Отримання даних з форми
		omega := r.FormValue("omega")
		tB := r.FormValue("tB")
		pM := r.FormValue("pM")
		kp := r.FormValue("kp")
		tM := r.FormValue("tM")
		zPerA := r.FormValue("zPerA")
		zPerP := r.FormValue("zPerP")
		fOmega, _ := strconv.ParseFloat(omega, 64)
		fTB, _ := strconv.ParseFloat(tB, 64)
		fPM, _ := strconv.ParseFloat(pM, 64)
		fKP, _ := strconv.ParseFloat(kp, 64)
		fTM, _ := strconv.ParseFloat(tM, 64)
		fZPerA, _ := strconv.ParseFloat(zPerA, 64)
		fZPerP, _ := strconv.ParseFloat(zPerP, 64)
		// Розрахунки
		mwNedA := fOmega * fTB * fPM * fTM
		mwNedP := fKP * fPM * fTM
		losses := fZPerA*mwNedA + fZPerP*mwNedP
		// Формування результату
		result := fmt.Sprintf(`
      Математичне очікування аварійного недопостачання: %.2f кВт-год
      Математичне очікування планового недопостачання: %.2f кВт-год
      Загальні втрати: %.2f грн
    `, mwNedA, mwNedP, losses)
		//Передача даних у шаблон
		data := PageData{
			Omega:  omega,
			TB:     tB,
			PM:     pM,
			KP:     kp,
			TM:     tM,
			ZPerA:  zPerA,
			ZPerP:  zPerP,
			Result: result,
		}
		tmpl.Execute(w, data)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	var err error
	// Завантаження HTML-шаблону
	tmpl, err = template.ParseFiles("template.html")
	if err != nil {
		log.Fatalf("Помилка парсингу шаблону: %v", err)
	}
	http.HandleFunc("/", handlerIndex)
	http.HandleFunc("/calculate", handlerCalculate)
	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
