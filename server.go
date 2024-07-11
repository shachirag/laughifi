package main

import (
	"laughifi/app"
	"laughifi/database"
	graph "laughifi/graph/resolvers"
	"laughifi/middleware"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
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

	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		KeepAlivePingInterval: 10 * time.Second,
	})

	srv.AddTransport(&transport.POST{})
	srv.AddTransport(&transport.Options{})

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
			// customer
			"LoginCustomer":             false,
			"ResendOtp":                 false,
			"ForgotPassword":            false,
			"ResetPassword":             false,
			"Signup":                    false,
			"VerifyOtpForSignup":        false,
			"SocialLoginCustomer":       false,
			"VerifyOtpForResetPassword": false,
			"ChangePassword":            true,
			"GetUserData":               true,
			"EditCustomer":              true,
			"GetLaughifiCustomers":      true,
			"GetTemplates":              true,
			"SendFriendRequest":         true,
			"GetFriendRequests":         true,
			"GetFriends":                true,
			"AcceptRejectRequest":       true,
			"CancelFriendRequest":       true,
			"FilledTemplates":           true,
			"GetFilledTemplates":        true,
			"FilledWouldYouRather":      true,
			"GetWouldYouRathers":        true,
			"GetFilledWouldRathers":     true,
			"TemplatePlayWithFriends":   true,
			"GetFriendTemplates":        true,

			// admin
			"LoginAdmin":                     false,
			"AdminForgotPassword":            false,
			"AdminVerifyOtpForResetPassword": false,
			"AdminResetPassword":             false,
			"AdminChangePassword":            true,
			"AdminEditAdmin":                 true,
			"GetAdminData":                   true,
			"AddCatgeory":                    true,
			"GetCategoryData":                true,
			"DeletedCategoryData":            true,
			"EditCategory":                   true,
			"GetCategories":                  true,
			"GetDashboardCounts":             true,
			"AddTemplate":                    true,
			"DeletedTemplateData":            true,
			"EditTemplate":                   true,
			"GetAllCategories":               true,
			"GetTemplateData":                true,
			"GetAllTemplates":                true,
			"AddWouldYouRather":              true,
			"DeleteWouldYouRather":           true,
			"GetWouldYouRatherData":          true,
			"EditWouldYouRather":             true,
			"GetAllWouldYouRathers":          true,
		}
		if authRequiredOperations[opName] {
			middleware.ValidateJWT(srv).ServeHTTP(w, r)
		} else {
			srv.ServeHTTP(w, r)
		}
	}))

	http.Handle("/schema", srv)

	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))

	log.Printf("connect to http://localhost:%s/ for GraphQL Playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
