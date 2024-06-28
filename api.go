package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/go-sql-driver/mysql"
)

type Pessoa struct {
	Id   string `json:"id"`
	Nome string `json:"nome"`
}

var db *sql.DB

func initDB() {
	var err error
	connStr := "username:password@tcp(localhost:3306)/jean"
	db, err = sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS pessoas (
		id INT AUTO_INCREMENT PRIMARY KEY,
		nome VARCHAR(255) NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

func getPessoas(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, nome FROM pessoas")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pessoas []Pessoa
	for rows.Next() {
		var p Pessoa
		if err := rows.Scan(&p.Id, &p.Nome); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pessoas = append(pessoas, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pessoas)
}

func getPessoaPorId(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var p Pessoa
	err := db.QueryRow("SELECT id, nome FROM pessoas WHERE id = ?", id).Scan(&p.Id, &p.Nome)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func criarPessoa(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var novaPessoa Pessoa
	err := json.NewDecoder(r.Body).Decode(&novaPessoa)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var id int
	err = db.QueryRow("INSERT INTO pessoas (nome) VALUES (?) RETURNING id", novaPessoa.Nome).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	novaPessoa.Id = fmt.Sprintf("%d", id)
	json.NewEncoder(w).Encode(novaPessoa)
}

func deletarPessoaPorId(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	_, err := db.Exec("DELETE FROM pessoas WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World")
}

func main() {
	initDB()
	defer db.Close()

	r := mux.NewRouter()

	r.HandleFunc("/", getRoot)
	r.HandleFunc("/pessoas", getPessoas).Methods(http.MethodGet)
	r.HandleFunc("/pessoas", criarPessoa).Methods(http.MethodPost)
	r.HandleFunc("/pessoas/{id}", getPessoaPorId).Methods(http.MethodGet)
	r.HandleFunc("/pessoas/{id}", deletarPessoaPorId).Methods(http.MethodDelete)

	fmt.Println("Servidor rodando na porta :3333")
	err := http.ListenAndServe(":3333", r)
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
	}
}
