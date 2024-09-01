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
		l := log.New(os.Stdout, "[server2]", log.Ldate|log.Ltime)
		l.Printf("running server 2")
		io.WriteString(w, "Hello world from Server 2\n")
	}
	http.HandleFunc("/", homeHandler)
	fmt.Println("Starting Server 2.... on port 8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}