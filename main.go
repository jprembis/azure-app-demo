package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

type record struct {
	Key   uint64 `json:"key"`
	Value []byte `json:"value"`
}

type memLog struct {
	mu      sync.Mutex
	records []record
}

type listing struct {
	Records []record `json:"records"`
}

type appendRequest struct {
	Record record `json:"record"`
}

type appendResponse struct {
	Key uint64 `json:"key"`
}

type getRequest struct {
	Key uint64 `json:"key"`
}

type getResponse struct {
	Record record `json:"record"`
}

var errKeyNotFound = fmt.Errorf("key not found")

func newMemLog() *memLog {
	return &memLog{}
}

func (l *memLog) append(r record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	r.Key = uint64(len(l.records)) // = index of r
	l.records = append(l.records, r)
	return r.Key, nil
}

func (l *memLog) get(key uint64) (record, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if key >= uint64(len(l.records)) {
		return record{}, errKeyNotFound
	}
	return l.records[key], nil
}

func (l *memLog) handleAppend(w http.ResponseWriter, r *http.Request) {
	var req appendRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	key, err := l.append(req.Record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := appendResponse{Key: key}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (l *memLog) handleGet(w http.ResponseWriter, r *http.Request) {
	var req getRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err == io.EOF {
		if len(l.records) > 0 {
			res := listing{Records: l.records}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(res)
		}
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	record, err := l.get(req.Key)
	if err == errKeyNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := getResponse{Record: record}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (l *memLog) indexHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		l.handleGet(w, r)
	case http.MethodPost:
		l.handleAppend(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method Not Allowed", 405)
	}
}

func main() {
	port, ok := os.LookupEnv("HTTP_PLATFORM_PORT")
	if !ok {
		port = "8080"
	}
	l := newMemLog()
	mux := http.NewServeMux()
	mux.HandleFunc("/", l.indexHandler)
	err := http.ListenAndServe(":"+port, mux)
	log.Fatal(err)
}
