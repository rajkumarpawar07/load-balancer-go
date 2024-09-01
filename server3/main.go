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
		l := log.New(os.Stdout, "[server3]", log.Ldate|log.Ltime)
		l.Printf("running server 3")
		io.WriteString(w, "Hello world from Server 3\n")
	}
	http.HandleFunc("/", homeHandler)
	fmt.Println("Starting Server 3.... on port 8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}