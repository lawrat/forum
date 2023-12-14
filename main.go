package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

const addr = "localhost"
const port = ":9090"

func renderTemplate(w http.ResponseWriter, tmpl string) {
	t, err := template.ParseFiles("./templates/" + tmpl + ".html")
	if err != nil {
		fmt.Fprint(w, "MODELE INTROUVABLE...")
	}
	t.Execute(w, nil)
}
func profil(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "profil")
}
func connexion(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		renderTemplate(w, "connexion")
	case "POST":
		db, err := sql.Open("sqlite3", "C:/sqlite/mabase.db")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer db.Close()

		// Récupérer les données du formulaire
		username := r.FormValue("username")
		motDePasse := r.FormValue("password")

		// Vérifier les informations de connexion
		var motDePasseDB string
		err = db.QueryRow("SELECT password_hash FROM users WHERE username=?", username).Scan(&motDePasseDB)
		if err != nil {
			fmt.Println(err)
			return
		}

		if motDePasse == motDePasseDB {
			fmt.Println("Connexion réussie !")
			http.Redirect(w, r, "/profil", http.StatusSeeOther)
			// Afficher le profil de l'utilisateur ici
		} else {
			fmt.Println("Mot de passe incorrect.")
			fmt.Fprint(w, "Mot de passe incorrect.")
		}
	}

}
func inscription(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		renderTemplate(w, "inscription")
	case "POST":
		nom := r.FormValue("nom")
		prenom := r.FormValue("prenom")
		email := r.FormValue("email")
		username := r.FormValue("username")
		password_hash := r.FormValue("password_hash")

		db, err := sql.Open("sqlite3", "C:/sqlite/mabase.db")
		if err != nil {
			log.Println("ERREUR LORS DE L'OUVERTURE DE LA BASE DE DONNEE:", err)
			http.Error(w, "Erreur lors de l'ouverture de la base de données", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Vérifier la connexion à la base de données
		err = db.Ping()
		if err != nil {
			log.Println("ERREUR LORS DE LA CONNEXION A LA BASE DE DONNEE:", err)
			http.Error(w, "Erreur lors de la connexion à la base de données", http.StatusInternalServerError)
			return
		}

		// Exécuter la requête d'insertion
		result, err := db.Exec(`INSERT INTO users (nom, prenom, email, username, password_hash)
							VALUES (?, ?, ?, ?, ?)`,
			nom, prenom, email, username, password_hash)
		if err != nil {
			fmt.Println("ERREUR LORS DE L'INSERTION DANS LA BASE DE DONNE")

		}

		// Récupérer l'ID de la ligne insérée
		id, err := result.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Nouvel utilisateur inséré avec l'ID: %d\n", id)
		fmt.Fprint(w, "Inscription réussie !")

		// http.Redirect(w, r, "http://localhost:9090/profil", http.StatusSeeOther)
	default:
		fmt.Fprint(w, "METHODE NON PRIS EN CHARGE")

	}
}
func main() {
	http.HandleFunc("/", connexion)
	http.HandleFunc("/inscription", inscription)
	http.HandleFunc("/profil", profil)

	fmt.Printf("serverur en cours d'execution sur http://%s%s\n", addr, port)
	http.ListenAndServe(port, nil)
}
