package main

import (
	"laughifi/app"
	"laughifi/database"
	graph "laughifi/graph/resolvers"
	"laughifi/middleware"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

func main() {

	err := app.SetupAndRunApp()
	if err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	db := database.Connect()
	s3 := database.GetS3Uploader()
	ses := database.GetSesClient()
	resolver := &graph.Resolver{
		DB:        db,
		S3Client:  s3,
		SESClient: ses,
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	})

	http.Handle("/query", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opName := r.Header.Get("X-GraphQL-Operation-Name")
		if opName == "" {
			http.Error(w, "Operation name header is required", http.StatusBadRequest)
			return
		}

		authRequiredOperations := map[string]bool{
			"ChangePassword":            true,
			"GetUserData":               true,
			"LoginCustomer":             false,
			"ForgotPassword":            false,
			"ResetPassword":             false,
			"Signup":                    false,
			"VerifyOtpForSignup":        false,
			"SocialLoginCustomer":       false,
			"VerifyOtpForResetPassword": false,
			"GetLaughifiCustomers":          true,
			"GetTemplates":              true,
			"SendFriendRequest":         true,
			"GetFriendRequests":         true,
			"GetFriends":                true,
			"AcceptRejectRequest":       true,
		}
		if authRequiredOperations[opName] {
			middleware.ValidateJWT(srv).ServeHTTP(w, r)
		} else {
			srv.ServeHTTP(w, r)
		}
	}))

	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))

	log.Printf("connect to http://localhost:%s/ for GraphQL Playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// package main

// import (
// 	"laughifi/config"
// 	"laughifi/database"
// 	graph "laughifi/graph/resolvers"
// 	"log"
// 	"net/http"
// 	"os"

// 	"github.com/99designs/gqlgen/graphql/handler"
// 	"github.com/99designs/gqlgen/graphql/playground"
// )

// const defaultPort = "8080"

// func main() {
// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = defaultPort
// 	}

// 	err := database.SetupAWSClient()
// 	if err != nil {
// 		panic(err)
// 	}

// 	err = config.LoadENV()
// 	if err != nil {
// 		panic(err)
// 	}

// 	db := database.Connect()
// 	s3 := database.GetS3Uploader()
// 	ses := database.GetSesClient()
// 	resolver := &graph.Resolver{
// 		DB:        db,
// 		S3Client:  s3,
// 		SESClient: ses,
// 	}
// 	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

// 	http.Handle("/query", srv)

// 	http.Handle("/", playground.Handler("GraphQL playground", "/query"))

// 	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
// 	log.Fatal(http.ListenAndServe(":"+port, nil))
// }
