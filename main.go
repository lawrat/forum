package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

const addr = "localhost"
const port = ":9090"

const smtpUsername = "salllawratou248@gmail.com"
const smtpPassword = "Mardi@2021"

var activeSessions = make(map[string]bool)

func renderTemplate(w http.ResponseWriter, tmpl string, errorMessage string) {
	t, err := template.ParseFiles("./templates/" + tmpl + ".html")
	if err != nil {
		fmt.Fprint(w, "MODELE INTROUVABLE...")
	}
	t.Execute(w, struct{ ErrorMessage string }{errorMessage})
}

func mdp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		renderTemplate(w, "mdp", "")
	case "POST":
		email := r.FormValue("email")

		db, err := sql.Open("sqlite3", "C:/sqlite/mabase.db")
		if err != nil {
			log.Println(err)
			http.Error(w, "Erreur lors de l'accès à la base de données", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		var emailDB string
		err = db.QueryRow("SELECT email FROM users WHERE email=?", email).Scan(&emailDB)
		if err != nil {
			log.Println(err)
			http.Error(w, "Erreur lors de la récupération des informations de l'utilisateur", http.StatusInternalServerError)
		}
		fmt.Println("Email récupéré avec succès")

		if email == emailDB {
			err := sendMail(email)
			if err != nil {
				log.Println("Erreur lors de l'envoi de l'e-mail:", err)
				renderTemplate(w, "mdp", "Erreur lors de l'envoi de l'e-mail")
				return
			}
			fmt.Println("E-mail envoyé avec succès")
		}
	}
}

func sendMail(email string) error {
	from := smtpUsername
	to := []string{email}
	subject := "Réinitialisation de mot de passe"
	body := "Cliquez sur le lien suivant pour réinitialiser votre mot de passe : http://localhost:9090/renitialiser"

	msg := "Subject: " + subject + "\n\n" + body

	auth := smtp.PlainAuth("", from, smtpPassword, "smtp.gmail.com")

	err := smtp.SendMail("smtp.gmail.com:587", auth, from, to, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}
func verification(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "verification", "")
}
func fil(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'identifiant de session à partir du cookie
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	sessionID := cookie.Value

	// Vérifier si la session est active
	if !activeSessions[sessionID] {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Récupérer les informations de l'utilisateur à partir de la base de données
	db, err := sql.Open("sqlite3", "C:/sqlite/mabase.db")
	if err != nil {
		log.Println(err)
		http.Error(w, "Erreur lors de l'accès à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var nom, prenom string
	err = db.QueryRow("SELECT nom, prenom FROM users WHERE username=?", sessionID).Scan(&nom, &prenom)
	if err != nil {
		if err == sql.ErrNoRows {
			// L'utilisateur n'a pas été trouvé dans la base de données
			log.Println("Utilisateur non trouvé dans la base de données")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		log.Println("Erreur lors de la récupération des informations de l'utilisateur:", err)
		http.Error(w, "Erreur lors de la récupération des informations de l'utilisateur", http.StatusInternalServerError)
		return
	}

	// Afficher le profil de l'utilisateur avec les informations récupérées
	t, err := template.ParseFiles("./templates/profil.html")
	if err != nil {
		fmt.Fprint(w, "MODELE INTROUVABLE...")
		return
	}
	t.Execute(w, struct{ Nom, Prenom string }{nom, prenom})
}

func profil(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "fil", "")
}
func renitialiser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		renderTemplate(w, "renitialiser", "")
	case "POST":
		username := r.FormValue("username")
		passwordHash := r.FormValue("password_hash")
		confirmPassword := r.FormValue("confirmPassword")

		if passwordHash != confirmPassword {
			renderTemplate(w, "inscription", "Les mots de passe ne correspondent pas.")
			return
		}

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
		var userDB, password_hashBD string
		err = db.QueryRow("SELECT username,password_hash FROM users WHERE username=?").Scan(&userDB, &password_hashBD)
		if err != nil {
			log.Println(err)
			http.Error(w, "Erreur lors de la récupération des informations de l'utilisateur", http.StatusInternalServerError)
			return
		}
		if username == userDB {
			_, err = db.Exec("UPDATE users SET password_hash=? WHERE username=?", passwordHash)
			if err != nil {
				log.Println(err)
				http.Error(w, "Erreur lors de la mise à jour des informations de l'utilisateur", http.StatusInternalServerError)
				return
			}
			// Rediriger vers la page de profil après la modification
			http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		}
	default:
		fmt.Fprint(w, "METHODE NON PRIS EN CHARGE")
	}

}

func connexion(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Vérifiez si l'utilisateur est déjà connecté
		cookie, err := r.Cookie("session")
		if err == nil {
			sessionID := cookie.Value
			if activeSessions[sessionID] {
				// L'utilisateur est déjà connecté, redirigez vers la page fil ou le profil de l'utilisateur
				http.Redirect(w, r, "/fil", http.StatusSeeOther)
				return
			}
		}
		renderTemplate(w, "connexion", "")
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
		var motDePasseDB, usernameDB string
		err = db.QueryRow("SELECT username, password_hash FROM users WHERE username=?", username).Scan(&usernameDB, &motDePasseDB)
		if err != nil {
			if err == sql.ErrNoRows {
				// Aucune ligne trouvée, c'est-à-dire utilisateur inexistant
				renderTemplate(w, "connexion", "Utilisateur inexistant, veuillez créer un compte.")
				return
			}
			fmt.Println(err)
			return
		}

		if username == usernameDB && motDePasse == motDePasseDB {
			// Générer un identifiant de session et définir un cookie
			sessionID := username // Utilisez l'username comme identifiant de session
			http.SetCookie(w, &http.Cookie{
				Name:  "session",
				Value: sessionID,
			})

			// Marquer la session comme active
			activeSessions[sessionID] = true

			fmt.Println("Connexion réussie !")
			http.Redirect(w, r, "/fil", http.StatusSeeOther)
			// Afficher le profil de l'utilisateur ici
		} else if username == "" {
			renderTemplate(w, "connexion", "Veuillez entrer votre nom utilisateur.")
		} else if motDePasse == "" {
			renderTemplate(w, "connexion", "Veuillez entrer votre mot de passe.")
		} else if username == usernameDB && motDePasse == "" {
			renderTemplate(w, "connexion", "Veuillez entrer votre mot de passe.")
		} else if username == "" && motDePasse == motDePasseDB {
			renderTemplate(w, "connexion", "Veuillez entrer votre nom utilisateur.")
		} else if username == "" && motDePasse == "" {
			renderTemplate(w, "connexion", "Veuillez entrer votre nom utilisateur et votre mot de passe.")
		} else if username == usernameDB && motDePasse != motDePasseDB {
			renderTemplate(w, "connexion", "Mot de passe incorrect, veuillez réessayer.")
		} else {
			fmt.Println("Mot de passe incorrect.")
			renderTemplate(w, "connexion", "Nom d'utilisateur ou mot de passe incorrect.")
		}
	}
}
func inscription(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		renderTemplate(w, "inscription", "")
	case "POST":
		nom := r.FormValue("nom")
		prenom := r.FormValue("prenom")
		email := r.FormValue("email")
		username := r.FormValue("username")
		passwordHash := r.FormValue("password_hash")
		confirmPassword := r.FormValue("confirmPassword")

		if passwordHash != confirmPassword {
			renderTemplate(w, "inscription", "Les mots de passe ne correspondent pas.")
		}

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
			nom, prenom, email, username, passwordHash)
		if err != nil {
			fmt.Println("ERREUR LORS DE L'INSERTION DANS LA BASE DE DONNE")

		}

		// Récupérer l'ID de la ligne insérée
		id, err := result.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Nouvel utilisateur inséré avec l'ID: %d\n", id)
		http.Redirect(w, r, "http://localhost:9090", http.StatusSeeOther)
	default:
		fmt.Fprint(w, "METHODE NON PRIS EN CHARGE")
	}
}

func deconnexion(w http.ResponseWriter, r *http.Request) {
	// Effacer le cookie de session
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionID := cookie.Value
		http.SetCookie(w, &http.Cookie{
			Name:   "session",
			Value:  "",
			MaxAge: -1,
		})

		// Marquer la session comme inactive
		activeSessions[sessionID] = false
	}

	// Rediriger vers la page de connexion
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func generateSessionID() string {
	// Implémentez une logique plus robuste pour générer un identifiant de session unique
	return "some_generated_session_id"
}

func main() {
	// Ajoutez une nouvelle route pour la déconnexion
	http.HandleFunc("/deconnexion", deconnexion)
	http.HandleFunc("/verification", verification)
	http.HandleFunc("/mdp", mdp)
	http.HandleFunc("/renitialiser", renitialiser)
	http.HandleFunc("/", connexion)
	http.HandleFunc("/fil", fil)
	http.HandleFunc("/inscription", inscription)
	http.HandleFunc("/profil", profil)
	fmt.Printf("Serveur en cours d'exécution sur http://%s%s\n", addr, port)
	http.ListenAndServe(port, nil)
}
