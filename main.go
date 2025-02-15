package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

// Kullanıcı yapısı
type Kullanici struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Film struct {
	ID        int    `json:"id"`
	Baslik    string `json:"baslik"`
	Aciklama  string `json:"aciklama"`
	Yonetmen  string `json:"yonetmen"`
	YayinYili int    `json:"yayin_yili"`
}

type Dizi struct {
	ID            int    `json:"id"`
	Baslik        string `json:"baslik"`
	Aciklama      string `json:"aciklama"`
	Kanal         string `json:"kanal"`
	BaslangicYili int    `json:"baslangic_yili"`
	BitisYili     int    `json:"bitis_yili"`
}

// Secret key for JWT
var jwtKey = []byte("my_secret_key")

func main() {
	var err error
	db, err = sql.Open("mysql", "root:Asdasd123123.@tcp(localhost:3306)/test")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := mux.NewRouter()

	// Routes for public access
	r.HandleFunc("/login", login).Methods("POST")
	r.HandleFunc("/filmler", filmleriGetir).Methods("GET")
	r.HandleFunc("/diziler", dizileriGetir).Methods("GET")

	// Protected routes
	r.HandleFunc("/filmler/{id}", filmiGuncelle).Methods("PUT")
	r.HandleFunc("/filmler/{id}", filmiSil).Methods("DELETE")
	r.HandleFunc("/diziler", diziEkle).Methods("POST")
	r.HandleFunc("/diziler/{id}", diziyiGuncelle).Methods("PUT")
	r.HandleFunc("/diziler/{id}", diziyiSil).Methods("DELETE")

	fmt.Println("Sunucu 8080 portunda çalışıyor...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func login(w http.ResponseWriter, r *http.Request) {
	var kullanici Kullanici
	// JSON'dan kullanıcı bilgilerini alıyoruz
	json.NewDecoder(r.Body).Decode(&kullanici)

	// Burada basit bir kontrol yapıyoruz (gerçek uygulamada şifre hash ile karşılaştırılmalı)
	if kullanici.Username != "admin" || kullanici.Password != "password" {
		http.Error(w, "Geçersiz kullanıcı adı veya şifre", http.StatusUnauthorized)
		return
	}

	// Token oluşturuyoruz
	token := jwt.New(jwt.SigningMethodHS256)

	// Claims kısmına kullanıcı adını ekliyoruz
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = kullanici.Username
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix() // Token geçerlilik süresi 1 saat

	// Token'ı imzalıyoruz
	t, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Token'ı döndürüyoruz
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": t,
	})
}

func filmleriGetir(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM filmler")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var filmler []Film
	for rows.Next() {
		var film Film
		err := rows.Scan(&film.ID, &film.Baslik, &film.Aciklama, &film.Yonetmen, &film.YayinYili)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		filmler = append(filmler, film)
	}

	json.NewEncoder(w).Encode(filmler)
}

func diziEkle(w http.ResponseWriter, r *http.Request) {
	// Token doğrulaması
	tokenString := r.Header.Get("Authorization")
	if !isValidToken(tokenString) {
		http.Error(w, "Geçersiz token", http.StatusUnauthorized)
		return
	}

	var dizi Dizi
	json.NewDecoder(r.Body).Decode(&dizi)

	_, err := db.Exec("INSERT INTO diziler (baslik, aciklama, kanal, baslangic_yili, bitis_yili) VALUES (?, ?, ?, ?, ?)", dizi.Baslik, dizi.Aciklama, dizi.Kanal, dizi.BaslangicYili, dizi.BitisYili)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Dizi eklendi!"))
}

func isValidToken(tokenString string) bool {
	if tokenString == "" {
		return false
	}

	// Bearer token'ı çıkarıyoruz
	tokenString = tokenString[7:]

	// Token'ı doğruluyoruz
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Geçersiz signing method")
		}
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return false
	}
	return true
}

func filmiGuncelle(w http.ResponseWriter, r *http.Request) {
	// Token doğrulaması
	tokenString := r.Header.Get("Authorization")
	if !isValidToken(tokenString) {
		http.Error(w, "Geçersiz token", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	var film Film
	json.NewDecoder(r.Body).Decode(&film)

	_, err := db.Exec("UPDATE filmler SET baslik=?, aciklama=?, yonetmen=?, yayin_yili=? WHERE id=?", film.Baslik, film.Aciklama, film.Yonetmen, film.YayinYili, params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Film güncellendi!"))
}

func filmiSil(w http.ResponseWriter, r *http.Request) {
	// Token doğrulaması
	tokenString := r.Header.Get("Authorization")
	if !isValidToken(tokenString) {
		http.Error(w, "Geçersiz token", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	_, err := db.Exec("DELETE FROM filmler WHERE id=?", params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Film silindi!"))
}

func diziyiGuncelle(w http.ResponseWriter, r *http.Request) {
	// Token doğrulaması
	tokenString := r.Header.Get("Authorization")
	if !isValidToken(tokenString) {
		http.Error(w, "Geçersiz token", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	var dizi Dizi
	json.NewDecoder(r.Body).Decode(&dizi)

	_, err := db.Exec("UPDATE diziler SET baslik=?, aciklama=?, kanal=?, baslangic_yili=?, bitis_yili=? WHERE id=?", dizi.Baslik, dizi.Aciklama, dizi.Kanal, dizi.BaslangicYili, dizi.BitisYili, params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Dizi güncellendi!"))
}

func diziyiSil(w http.ResponseWriter, r *http.Request) {
	// Token doğrulaması
	tokenString := r.Header.Get("Authorization")
	if !isValidToken(tokenString) {
		http.Error(w, "Geçersiz token", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	_, err := db.Exec("DELETE FROM diziler WHERE id=?", params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Dizi silindi!"))
}

func dizileriGetir(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM diziler")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var diziler []Dizi
	for rows.Next() {
		var dizi Dizi
		err := rows.Scan(&dizi.ID, &dizi.Baslik, &dizi.Aciklama, &dizi.Kanal, &dizi.BaslangicYili, &dizi.BitisYili)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		diziler = append(diziler, dizi)
	}

	json.NewEncoder(w).Encode(diziler)
}
