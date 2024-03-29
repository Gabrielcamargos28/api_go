package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type Pessoa struct {
	Id   string  `json:"id"`
	Nome string `json:"nome"`
}

var banco []Pessoa

func getPessoas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(banco)
}

func getPessoaPorId(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    for _, item := range banco {
        if item.Id == params["id"] {
            json.NewEncoder(w).Encode(item)
            return
        }
    }
    json.NewEncoder(w).Encode(&Pessoa{})
}


func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World")
}

func criarPessoa(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var novaPessoa Pessoa
	err := json.NewDecoder(r.Body).Decode(&novaPessoa)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	banco = append(banco, novaPessoa)
	json.NewEncoder(w).Encode(novaPessoa)
}
func getDeletePorId(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
    for index, item := range banco {
        if item.Id == params["id"] {
            banco = append(banco[:index], banco[index+1:]...)
            break
        }
        json.NewEncoder(w).Encode(banco)
    }
}
func main() {

	banco = []Pessoa{
		{Id: "1", Nome: "Jean"},
		{Id: "2", Nome: "Gabriel"},
	}

	r := mux.NewRouter()

	r.HandleFunc("/", getRoot)
	r.HandleFunc("/pessoas", getPessoas).Methods(http.MethodGet)
	r.HandleFunc("/pessoas", criarPessoa).Methods(http.MethodPost)
	r.HandleFunc("/pessoas/{id}", getPessoaPorId).Methods(http.MethodGet)
	r.HandleFunc("/pessoas/{id}", getDeletePorId).Methods(http.MethodDelete)
	fmt.Println("Servidor rodando na porta :3333")
	err := http.ListenAndServe(":3333", r)
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
	}
}