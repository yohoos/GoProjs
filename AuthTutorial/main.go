package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"encoding/json"
	"github.com/gorilla/handlers"
	"os"
	"github.com/dgrijalva/jwt-go"
	"time"
	"fmt"
)

func main() {

	type Product struct {
		Id          int
		Name        string
		Slug        string
		Description string
	}

	products := []Product{
		{Id: 1, Name: "Hover Shooters", Slug: "hover-shooters", Description: "Shoot your way to the top on 14 different hoverboards"},
		{Id: 2, Name: "Ocean Explorer", Slug: "ocean-explorer", Description: "Explore the depths of the sea in this one of a kind underwater experience"},
		{Id: 3, Name: "Dinosaur Park", Slug: "dinosaur-park", Description: "Go back 65 million years in the past and ride a T-Rex"},
		{Id: 4, Name: "Cars VR", Slug: "cars-vr", Description: "Get behind the wheel of the fastest cars in the world."},
		{Id: 5, Name: "Robin Hood", Slug: "robin-hood", Description: "Pick up the bow and arrow and master the art of archery"},
		{Id: 6, Name: "Real World VR", Slug: "real-world-vr", Description: "Explore the seven wonders of the world in VR"},
	}

	StatusHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API is up and running"))
	})

	ProductsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := json.Marshal(products)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(payload))
	})

	AddFeedbackHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var product Product
		vars := mux.Vars(r)
		slug := vars["slug"]

		for _, p := range products {
			if p.Slug == slug {
				product = p
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if product.Slug != "" {
			payload, _ := json.Marshal(product)
			w.Write([]byte(payload))
		} else {
			w.Write([]byte("Product Not Found"))
		}
	})

	//jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
	//	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
	//		return mySigningKey, nil
	//	},
	//	SigningMethod: jwt.SigningMethodHS256,
	//})

	validate := func(protectedPage http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.Cookies()[0])
			cookie, err := r.Cookie("Authorization")
			if err != nil {
				http.NotFound(w, r)
				return
			}

			token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				return []byte("secret"), nil
			})
			if err != nil {
				http.NotFound(w, r)
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				fmt.Println(claims["admin"], claims["name"])
			} else {
				fmt.Println(err)
			}

			protectedPage(w, r)
		})
	}

	r := mux.NewRouter()

	r.Handle("/", http.FileServer(http.Dir("./views/")))

	r.Handle("/status", StatusHandler).Methods("GET")

	r.Handle("/products", validate(ProductsHandler)).Methods("GET")

	r.Handle("/token", GetTokenHandler).Methods("GET")

	r.Handle("/products/{slug}/feedback", validate(AddFeedbackHandler)).Methods("POST")

	r.PathPrefix("/static/").
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	http.ListenAndServe(":3000", handlers.LoggingHandler(os.Stdout, r))
}

var NotImplemented = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Not Implemented"))
})

var mySigningKey = []byte("secret")

var GetTokenHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["admin"] = true
	claims["name"] = "Ado Kukic"
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, _ := token.SignedString(mySigningKey)

	cookie := &http.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)

	fmt.Println(cookie)

	http.Redirect(w, r, "/", 307)
})
