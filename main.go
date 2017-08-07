package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"
)

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the application")

	u := newUcenter()

	//利用cpu多核来处理请求
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.Handle("/room", u)

	log.Println("Starting web server on ", *addr)

	go u.run()
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("Listen and server :", err.Error())
	}

}
