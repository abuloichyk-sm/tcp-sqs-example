package main

func main() {

	// var b bytes.Buffer
	// log.SetOutput(&b)
	// log.SetFlags(log.Lmicroseconds)

	// go func() {
	// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 		fmt.Fprintf(w, "Hello! Logs from switcher: \n%s", b.String())
	// 	})

	// 	http.ListenAndServe(":8083", nil)
	// }()

	sw := NewSwitcher(80, 100)

	sw.Run()
}

// hello world for test connection
// func main() {
// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
// 	})

// 	http.ListenAndServe(":8081", nil)
// }
