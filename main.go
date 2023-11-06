package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

type Record struct {
	Key   uint64 `json:"key"`
	Value []byte `json:"value"`
}

type Log struct {
	mu      sync.Mutex
	records []Record
}

func NewLog() *Log {
	return &Log{}
}

func (l *Log) Write(r Record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	r.Key = uint64(len(l.records)) // = index of r
	l.records = append(l.records, r)
	return r.Key, nil
}

var ErrKeyNotFound = fmt.Errorf("key not found")

func (l *Log) Read(key uint64) (Record, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if key >= uint64(len(l.records)) {
		return Record{}, ErrKeyNotFound
	}
	return l.records[key], nil
}

type RequestFromWriter struct {
	Record Record `json:"record"`
}
type ResponseToWriter struct {
	Key uint64 `json:"key"`
}
type RequestFromReader struct {
	Key uint64 `json:"key"`
}
type ResponseToReader struct {
	Record Record `json:"record"`
}

func (l *Log) handleWrite(w http.ResponseWriter, r *http.Request) {
	var req RequestFromWriter
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	key, err := l.Write(req.Record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := ResponseToWriter{Key: key}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (l *Log) handleRead(w http.ResponseWriter, r *http.Request) {
	var req RequestFromReader
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	record, err := l.Read(req.Key)
	if err == ErrKeyNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := ResponseToReader{Record: record}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (l *Log) indexHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		l.handleRead(w, r)
	case http.MethodPost:
		l.handleWrite(w, r)
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
	l := NewLog()
	mux := http.NewServeMux()
	mux.HandleFunc("/", l.indexHandler)
	err := http.ListenAndServe(":"+port, mux)
	log.Fatal(err)
}
