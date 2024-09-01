package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main(){

	homeHandler := func (w http.ResponseWriter, r *http.Request)  {
		l := log.New(os.Stdout, "[server1]", log.Ldate|log.Ltime)
		l.Printf("running server 1")
		io.WriteString(w, "Hello world from Server 1\n")
	}
	http.HandleFunc("/", homeHandler)
	fmt.Println("Starting Server 1.... on port 8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}