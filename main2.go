package main

import "net/http"

func main() {
	for i := 0; i < 100; i++ {
		go func() {
			client := http.Client{}
			http.NewRequest("GET", "http://localhost:8080/reservation", nil)

			client.Do()
		}()
	}
}
