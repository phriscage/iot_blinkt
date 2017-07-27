// Package main provides sample webserver that interfaces with the GPIO
// of Raspberry PIs via the sys
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	. "github.com/alexellis/blinkt_go/sysfs"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

type ResponseCore struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
}

type ResponseData struct {
	ResponseCore
	Data map[string]interface{} `json:"data,omitempty"`
}

type ResponseError struct {
	ResponseCore
	Errors []string `json:"errors,omitempty"`
}

type Color struct {
	r, g, b int
}

// Custom Response Error Handlers with errors array
func ResponseErrorHandler(w http.ResponseWriter, r *http.Request, code int, errors []string) {
	if len(errors) != 0 {
		log.Print(errors)
	}
	responseData := &ResponseError{ResponseCore{code, http.StatusText(code)}, errors}
	body, err := json.Marshal(responseData)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
		return
	}
	// Always set Headers before Writing them
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(body))
}

func ResponseHandler(w http.ResponseWriter, r *http.Request, code int, data map[string]interface{}) {
	responseData := &ResponseData{ResponseCore{code, http.StatusText(code)}, data}
	body, err := json.Marshal(responseData)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
		return
	}
	// Always set Headers before Writing them
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(body))
}

func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	ResponseErrorHandler(w, r, http.StatusMethodNotAllowed, nil)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	ResponseHandler(w, r, http.StatusNotFound, nil)
}

// Main Handlers

// Hello Handler checks a parameter and returns the response
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if name == "" {
		name = "Name does not exist"
	}
	data := make(map[string]interface{})
	sites := []string{"a", "b", "c"}
	data["method"] = r.Method
	data["url"] = fmt.Sprintf("%s", r.URL)
	data["sites"] = sites
	data["name"] = name
	ResponseHandler(w, r, http.StatusOK, data)
}

// BlinktsHandler takes color and delay params and maps them to appropriate
// rgb color schemes for blinkt. Then turns on a random LED interface for the
// specified delay
func BlinktsHandler(w http.ResponseWriter, r *http.Request) {
	// Default actions for the resource
	actions := map[string]bool{
		"random": true,
		"all":    true,
		"pixel":  true,
	}
	// Default length of pixels. Need Blinkt class struct in capital
	pixelLength := 8
	var pixelId int

	vars := mux.Vars(r)
	action := vars["action"]

	if !actions[action] {
		ResponseHandler(w, r, http.StatusNotFound, nil)
		return
	}

	if action == "pixel" {
		id, err := strconv.Atoi(vars["id"])
		if err != nil || id >= pixelLength {
			log.Printf("'%s' is not a valid pixel", vars["id"])
			ResponseHandler(w, r, http.StatusNotFound, nil)
			return
		}
		pixelId = id
	}

	color := r.URL.Query().Get("color")
	// Default to blue for now
	if color == "" {
		color = "blue"
	}
	delay, err := strconv.Atoi(r.URL.Query().Get("delay"))
	// Default to 1000 for now
	if err != nil {
		log.Printf("%s", err)
	}
	if delay == 0 || delay > 5000 {
		delay = 1000
	}

	rgb := &Color{}
	if color == "blue" {
		rgb = &Color{0, 0, 255}
	} else if color == "red" {
		rgb = &Color{255, 0, 0}
	} else if color == "green" {
		rgb = &Color{0, 255, 0}
	}

	data := make(map[string]interface{})
	data["color"] = color
	data["delay"] = delay

	brightness := 0.5
	blinkt := NewBlinkt(brightness)
	//blinkt.SetClearOnExit(true)
	blinkt.Setup()
	blinkt.Clear()
	if action == "all" {
		log.Printf("Turning on all LEDs: '%s'", color)
		blinkt.SetAll(rgb.r, rgb.g, rgb.b)
		data["position"] = "all"
	} else if action == "pixel" {
		log.Printf("Turning LED [%d]: '%s'", pixelId, color)
		blinkt.SetPixel(pixelId, rgb.r, rgb.g, rgb.b)
		data["position"] = pixelId
	} else {
		random := rand.Intn(8)
		log.Printf("Turning LED [%d]: '%s'", random, color)
		blinkt.SetPixel(random, rgb.r, rgb.g, rgb.b)
		data["position"] = random
	}
	// Need to show twice for now...
	blinkt.Show()
	blinkt.Show()
	//log.Printf("Running...")
	Delay(delay)

	log.Printf("Turning off LEDs: 'off'")

	// Need to clear & show twice
	blinkt.Clear()
	blinkt.Show()
	blinkt.Show()

	ResponseHandler(w, r, http.StatusOK, data)
}

// Main router logic
func main() {
	addr := flag.String("addr", ":8080", "http listen address")
	flag.Parse()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/hello", HelloHandler).Methods("GET")
	router.HandleFunc("/hello", MethodNotAllowedHandler)
	router.HandleFunc("/hello/{name}", HelloHandler).Methods("GET")
	router.HandleFunc("/hello/{name}", MethodNotAllowedHandler)
	router.HandleFunc("/blinkts/{action}", BlinktsHandler).Methods("POST")
	router.HandleFunc("/blinkts/{action}", MethodNotAllowedHandler)
	router.HandleFunc("/blinkts/{action}/{id}", BlinktsHandler).Methods("POST")
	router.HandleFunc("/blinkts/{action}/{id}", MethodNotAllowedHandler)
	//router.Handle("/blinkts/random", handlers.MethodHandler{
	//"POST": http.HandlerFunc(BlinktsHandler),
	//})
	router.NotFoundHandler = http.HandlerFunc(NotFoundHandler)
	log.Printf("Starting web server on " + *addr)
	log.Fatal(http.ListenAndServe(*addr, handlers.CombinedLoggingHandler(os.Stderr, router)))
}
