const express = require("express");
const sqlite3 = require("sqlite3").verbose();
const bodyParser = require("body-parser");
const cookieParser = require("cookie-parser");

const app = express();
const port = 3000;

app.use(bodyParser.urlencoded({ extended: true }));
app.use(cookieParser());

const dbPath = "C:/sqlite/mabase.db";

app.get("/", (req, res) => {
  // Vérifiez si l'utilisateur est connecté (en fonction d'un cookie, d'une session, etc.)
  const isUserLoggedIn = req.cookies.isUserLoggedIn === "true";

  if (isUserLoggedIn) {
    res.redirect("/profil");
  } else {
    res.redirect("/connexion");
  }
});

app.get("/connexion", (req, res) => {
  res.sendFile(__dirname + "/connexion.html");
});

app.post("/connexion", (req, res) => {
  const db = new sqlite3.Database(dbPath);

  const username = req.body.username;
  const motDePasse = req.body.password;

  db.get(
    "SELECT password_hash FROM users WHERE username = ?",
    [username],
    (err, row) => {
      if (err) {
        console.error(err);
        res.status(500).send("Erreur lors de la connexion.");
        return;
      }

      if (row && motDePasse === row.password_hash) {
        console.log("Connexion réussie !");

        // Définir un cookie pour suivre l'état de connexion
        res.cookie("isUserLoggedIn", "true");
        app.get("/profil", (req, res) => {
          res.sendFile(__dirname + "/profil.html");
        });
        res.redirect("/profil");
      } else {
        console.log("Mot de passe incorrect.");
        const create = document.createElement("p");
        create.textContent = "mot de passe incorect";
        body.append(create);
      }
    }
  );

  db.close();
});

app.get("/inscription", (req, res) => {
  res.sendFile(__dirname + "/inscription.html");
});

app.post("/inscription", (req, res) => {
  const db = new sqlite3.Database(dbPath);

  const { nom, prenom, email, username, password_hash } = req.body;

  const query = `INSERT INTO users (nom, prenom, email, username, password_hash) VALUES (?, ?, ?, ?, ?)`;

  db.run(query, [nom, prenom, email, username, password_hash], function (err) {
    if (err) {
      return res
        .status(500)
        .send(
          "Erreur lors de l'insertion des données dans la base de données."
        );
    }
    console.log(`Nouvel utilisateur inséré avec l'ID ${this.lastID}`);
    res.redirect("/connexion");
  });

  db.close();
});

app.listen(port, () => {
  console.log(`Serveur démarré sur le port ${port}`);
});
