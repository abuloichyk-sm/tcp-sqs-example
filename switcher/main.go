package main

func main() {
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
