package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func initRouter() {
	router := mux.NewRouter()
	router.HandleFunc("/tricks/{id}", handleTricksPut).Methods("PUT")
	router.HandleFunc("/tricks", handleTricks).Methods("GET", "POST")
	http.ListenAndServe(":8080", router)
}

func handleTricks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleTricksGet(w, r)
	case http.MethodPost:
		handleTricksPost(w, r)
	case http.MethodPut:
		handleTricksPut(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func handleTricksGet(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM tricks")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tricks := make([]Trick, 0)
	for rows.Next() {
		var t Trick
		if err := rows.Scan(&t.ID, &t.Name, &t.TranslatedName, &t.Description, &t.Difficulty, &t.Progress); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		tricks = append(tricks, t)
	}

	if err := json.NewEncoder(w).Encode(tricks); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func handleTricksPost(w http.ResponseWriter, r *http.Request) {
	var t Trick
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	res, err := db.Exec("INSERT INTO tricks (name,translatedName,description,difficulty,progress) VALUES (?,?,?,?,?)", t.Name, &t.TranslatedName, t.Description, t.Difficulty, t.Progress)
	if err != nil {
		fmt.Println("Error: ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	t.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(t); err != nil {
		fmt.Println("Error: ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func handleTricksPut(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid Trick ID", http.StatusBadRequest)
		return
	}

	var t Trick
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	t.ID = id
	stmt, err := db.Prepare("UPDATE tricks SET name=?, translatedName=?, description=?, difficulty=?, progress=? WHERE id=?")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.Name, t.TranslatedName, t.Description, t.Difficulty, t.Progress, id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(t); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
