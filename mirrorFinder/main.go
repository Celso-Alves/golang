package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"workspace/mirrorFinder/mirrors"
)

type response struct {
	FastestURL string        `json:"fastest_url"`
	Latency    time.Duration `json:"latency"`
}

/*
A função findFastest é utulizada para fazer as requisições a todos os mirrors e calcular o
mais rápido de todos. Para fazer isso, em vez de fazer chamadas de API sequenciais para cada URL,
uma após a outra, usamos rotinas Go para solicitar paralelamente as URLs e,
quando uma goroutine retorna, paramos aí e retornamos os dados.
*/
func findFastest(urls [30]string) response {
	urlChan := make(chan string)
	latencyChan := make(chan time.Duration)
	/*
		A lógica qui é que, sempre que uma goroutine recebe uma resposta, ela grava dados em dois canais
		com a URL e as informações de latência, respectivamente. Ao receber os dados, os dois canais fazem
		a estrutura de resposta e retornam de a função findFastest.
		Quando essa função é retornada, todas as goroutines geradas são interrompidas de
		qualquer coisa que estejam fazendo. Assim, teremos a URL mais curta em urlChan e a menor
		latência em latencyChan.
	*/
	for _, url := range urls {
		mirrorURL := url
		go func() {
			start := time.Now()
			_, err := http.Get(mirrorURL + "/README")
			latency := time.Now().Sub(start) / time.Millisecond
			if err == nil {
				urlChan <- mirrorURL
				latencyChan <- latency
			}
		}()
	}
	return response{<-urlChan, <-latencyChan}
}

func main() {
	http.HandleFunc("/fastest-mirror", func(w http.ResponseWriter,
		r *http.Request) {
		response := findFastest(mirrors.MirrorList)
		respJSON, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(respJSON)
		if err != nil {
			return
		}
	})
	port := ":8000"
	server := &http.Server{
		Addr:           port,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Printf("Iniciando servidor na porta %sn", port)
	log.Fatal(server.ListenAndServe())
}
