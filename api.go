package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Pessoa struct {
	ID   int    `json:"id"`
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
		if err := rows.Scan(&p.ID, &p.Nome); err != nil {
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
	err := db.QueryRow("SELECT id, nome FROM pessoas WHERE id = ?", id).Scan(&p.ID, &p.Nome)
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

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var novaPessoa Pessoa
	err := json.NewDecoder(r.Body).Decode(&novaPessoa)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := db.Exec("INSERT INTO pessoas (nome) VALUES (?)", novaPessoa.Nome)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	novaPessoa.ID = int(id)
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

func atualizarPessoaPorId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var pessoaAtualizada Pessoa
	err = json.NewDecoder(r.Body).Decode(&pessoaAtualizada)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := db.Exec("UPDATE pessoas SET nome = ? WHERE id = ?", pessoaAtualizada.Nome, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.NotFound(w, r)
		return
	}

	pessoaAtualizada.ID = id
	json.NewEncoder(w).Encode(pessoaAtualizada)
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
	r.HandleFunc("/pessoas/{id}", atualizarPessoaPorId).Methods(http.MethodPut)

	fmt.Println("Servidor rodando na porta :3333")
	err := http.ListenAndServe(":3333", r)
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
	}
}
