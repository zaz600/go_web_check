// web_check.go
package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"
)

var (
	check_history []string
	url           string
	hist_length   int
	timeout       int
	ip            string
)

func main() {
	if !parse_args() {
		return
	}
	go check_loop()
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(ip, nil)
}

func parse_args() bool {
	flag.StringVar(&url, "url", "", "Адрес для проверки. Например, http://golang.org/")
	flag.IntVar(&timeout, "t", 30, "Период проверки в секундах. Должен быть больше 15 сек")
	flag.IntVar(&hist_length, "l", 30, "Длина истории в браузере")
	flag.StringVar(&ip, "i", ":8090", "ip:port для веб-статистики")
	flag.Parse()

	if url == "" {
		fmt.Println("Не задан параметр -url", url)
		return false
	}
	if timeout < 15 {
		fmt.Println("Значение -i должно быть больше 15. Задано: ", timeout)
		return false
	}
	if hist_length < 1 {
		fmt.Println("Значение -l должно быть больше 1. Задано: ", hist_length)
		return false
	}
	return true
}

func check_loop() {
	for {
		tm := time.Now().Format("2006-01-02 15:04:05")
		fmt.Println("Проверяем адрес ", url)
		// статус, который возвращает check, пока не используем, поэтому ставим _
		_, msg := check(url)
		log_to_file(tm, msg)
		save_history(tm, msg)
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
	if resp.StatusCode != 200 {
		return false, fmt.Sprintf("Ошибка. http-статус: %s", resp.StatusCode)
	}
	return true, fmt.Sprintf("Онлайн. http-статус: %d", resp.StatusCode)
}

func log_to_file(tm, s string) {
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

func save_history(tm, s string) {
	//  добавляет запись в массив с историей проверок
	check_history = append(check_history, fmt.Sprintf("%s %s", tm, s))
	if len(check_history) > hist_length {
		check_history = check_history[1:]
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	//  Выдает историю проверок в браузер
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, map[string]interface{}{"check_history": check_history, "url": url})
}
