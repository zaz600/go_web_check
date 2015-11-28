// web_check.go
package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	checkHistory []string
	url          string
	histLength   int
	timeout      int
	ip           string
	statesOk     string
	statesOkMap  map[string]bool
)

func main() {
	if !parseArgs() {
		return
	}
	go checkLoop()
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(ip, nil)
}

func parseArgs() bool {
	flag.StringVar(&url, "url", "", "Адрес для проверки. Например, http://golang.org/")
	flag.IntVar(&timeout, "t", 30, "Период проверки в секундах. Должен быть больше 15 сек")
	flag.IntVar(&histLength, "l", 30, "Длина истории в браузере")
	flag.StringVar(&ip, "i", ":8090", "ip:port для веб-статистики")
	flag.StringVar(&statesOk, "states", "200", "статусы, которые не считаются ошибочными. Например, 200,500")
	flag.Parse()

	if url == "" {
		fmt.Println("Не задан параметр -url", url)
		return false
	}
	if timeout < 15 {
		fmt.Println("Значение -i должно быть больше 15. Задано: ", timeout)
		return false
	}
	if histLength < 1 {
		fmt.Println("Значение -l должно быть больше 1. Задано: ", histLength)
		return false
	}
	statesOkMap = make(map[string]bool)
	for _, status := range strings.Split(statesOk, ",") {
		statesOkMap[status] = true
	}
	return true
}

func checkLoop() {
	for {
		tm := time.Now().Format("2006-01-02 15:04:05")
		fmt.Println("Проверяем адрес ", url)
		// статус, который возвращает check, пока не используем, поэтому ставим _
		_, msg := check(url)
		logToFile(tm, msg)
		saveHistory(tm, msg)
		fmt.Println(tm, msg)
		time.Sleep(time.Duration(timeout) * time.Second)
	}
}

func check(url string) (bool, string) {
	// возвращает true - если сервис доступен, false, если нет и текст сообщения
	resp, err := http.Get(url)

	if err != nil {
		return false, fmt.Sprintf("Ошибка соединения. %s", err)
	}

	defer resp.Body.Close()
	if _, ok := statesOkMap[strconv.Itoa(resp.StatusCode)]; ok != true {
		return false, fmt.Sprintf("Ошибка. http-статус: %d", resp.StatusCode)
	}
	return true, fmt.Sprintf("Онлайн. http-статус: %d", resp.StatusCode)
}

func logToFile(tm, s string) {
	//  Сохраняет сообщения в файл
	f, err := os.OpenFile("web_check.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(tm, err)
		return
	}
	defer f.Close()
	if _, err = f.WriteString(fmt.Sprintln(tm, s)); err != nil {
		fmt.Println(tm, err)
	}
}

func saveHistory(tm, s string) {
	//  добавляет запись в массив с историей проверок
	checkHistory = append(checkHistory, fmt.Sprintf("%s %s", tm, s))
	if len(checkHistory) > histLength {
		checkHistory = checkHistory[1:]
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	//  Выдает историю проверок в браузер
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, map[string]interface{}{"checkHistory": checkHistory, "url": url})
}
