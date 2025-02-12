package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

var tmpl *template.Template

func main() {
	var err error
	//Завантаження HTML з файлу
	tmpl, err = template.ParseFiles("template.html")
	if err != nil {
		fmt.Println("Помилка завантаження шаблону:", err)
		return
	}
	//Маршрути для обробки запитів
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	})
	//Обробник для завдання 1
	http.HandleFunc("/calculate1", calculateTask1)
	//Обробник для завдання 2
	http.HandleFunc("/calculate2", calculateTask2)

	fmt.Println("Сервер запущено на http://localhost:8080")
	http.ListenAndServe(":8080", nil) //Запуск сервера
}

// Перетворення рядка в число
func checkAndToDouble(input string) (float64, error) {
	return strconv.ParseFloat(input, 64)
}

// Завдання 1
func calculateTask1(w http.ResponseWriter, r *http.Request) {
	// Зчитування вхідних даних з форми
	hp, _ := checkAndToDouble(r.FormValue("hp"))
	cp, _ := checkAndToDouble(r.FormValue("cp"))
	sp, _ := checkAndToDouble(r.FormValue("sp"))
	np, _ := checkAndToDouble(r.FormValue("np"))
	op, _ := checkAndToDouble(r.FormValue("op"))
	wp, _ := checkAndToDouble(r.FormValue("wp"))
	ap, _ := checkAndToDouble(r.FormValue("ap"))

	// Коєфіцієнт для розрахунку складу сухої маси
	kpc := 100 / (100 - wp)

	// Коєфіцієнт для розрахунку складу горючої маси
	krg := 100 / (100 - wp - ap)

	//Розрахунок складу сухої маси
	hc, cc, sc, nc, oc, ac := hp*kpc, cp*kpc, sp*kpc, np*kpc, op*kpc, ap*kpc

	//Розрахунок складу горючої маси
	hg, cg, sg, ng, og := hp*krg, cp*krg, sp*krg, np*krg, op*krg

	// Нижча теплота згорання для робочої маси
	qph := (339*cp + 1030*hp - 108.8*(op-sp) - 25*wp) / 1000

	// Нижча теплота згорання для сухої маси
	qch := (qph + 0.025*wp) * 100 / (100 - wp)

	// Нижча теплота згорання для горючої маси
	qgh := (qph + 0.025*wp) * 100 / (100 - wp - ap)

	// Формування текстового результату
	result := fmt.Sprintf(`
Коефіцієнт переходу від робочої до сухої маси: %.3f
Коефіцієнт переходу від робочої до горючої маси: %.3f

Склад сухої маси:
Hc = %.3f %%
Cc = %.3f %%
Sc = %.3f %%
Nc = %.3f %%
Oc = %.3f %%
Ac = %.3f %%

Склад горючої маси:
Hg = %.3f %%
Cg = %.3f %%
Sg = %.3f %%
Ng = %.3f %%
Og = %.3f %%

Теплота згорання робочої маси: %.3f МДж/кг
Теплота згорання сухої маси: %.3f МДж/кг
Теплота згорання горючої маси: %.3f МДж/кг
`, kpc, krg, hc, cc, sc, nc, oc, ac, hg, cg, sg, ng, og, qph, qch, qgh)

	// Передача результата у шаблон та його відображення
	tmpl.Execute(w, map[string]string{"Result": result})
}

// Завдання 2
func calculateTask2(w http.ResponseWriter, r *http.Request) {
	// Зчитування вхідних даних з форми
	cg, _ := checkAndToDouble(r.FormValue("cg"))
	hg, _ := checkAndToDouble(r.FormValue("hg"))
	og, _ := checkAndToDouble(r.FormValue("og"))
	sg, _ := checkAndToDouble(r.FormValue("sg"))
	qi, _ := checkAndToDouble(r.FormValue("qi"))
	vg, _ := checkAndToDouble(r.FormValue("vg"))
	wg, _ := checkAndToDouble(r.FormValue("wg"))
	ag, _ := checkAndToDouble(r.FormValue("ag"))

	//Перерахунок елементарного складумазуту на робочу масу
	cp := cg * (100 - wg - ag) / 100.0
	hp := hg * (100 - wg - ag) / 100.0
	op := og * (100 - wg - ag) / 100.0
	sp := sg * (100 - wg - ag) / 100.0
	ap := ag * (100 - wg) / 100.0
	vp := vg * (100 - wg) / 100.0

	//Перерахунок  нижчої теплоти згоряння мазуту на робочу масу
	qri := qi*(100-wg-ap)/100 - 0.025*wg

	// Формування текстового результату
	result := fmt.Sprintf(`
Перерахунок елементарного складу мазуту на робочу масу:
Cp = %.3f %%
Hp = %.3f %%
Op = %.3f %%
Sp = %.3f %%
Ap = %.3f %%
Vp = %.3f мг/кг

Нижча теплота згоряння мазуту на робочу масу: %.3f МДж/кг
`, cp, hp, op, sp, ap, vp, qri)
	// Передача результата у шаблон та його відображення
	tmpl.Execute(w, map[string]string{"Result": result})
}
