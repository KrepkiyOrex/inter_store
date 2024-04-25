package others

import (
	"fmt"
	"net/http"
	"time"
)

// Test is main page
func GreetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Main page %s", time.Now())
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello to the north!")
}

func HeadersHandler(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}
