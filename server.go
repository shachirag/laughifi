package main

import (
	"laughifi/app"
	"laughifi/database"
	graph "laughifi/graph/resolvers"
	"laughifi/handlers/websocket"
	"laughifi/middleware"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/handlers"
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
	apiClient := database.GetApiClient()
	messagingClient := database.GetFirebaseMessagingClient()
	resolver := &graph.Resolver{
		DB:        db,
		S3Client:  s3,
		SESClient: ses,
		ApiClient: apiClient,
		MessagingClient: messagingClient,
	}
	

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	srv.AddTransport(&transport.POST{})
	srv.AddTransport(&transport.Options{})

	router := http.NewServeMux()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	})

	router.HandleFunc("/connect", websocket.WebsocketConnect(db))
	router.HandleFunc("/disconnect", websocket.WebsocketDisconnect(db))

	router.Handle("/query", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opName := r.Header.Get("X-GraphQL-Operation-Name")
		if opName == "" {
			http.Error(w, "Operation name header is required", http.StatusBadRequest)
			return
		}

		authRequiredOperations := mergeOperations(
			getAdminOperations(),
			getCustomerOperations(),
		)

		if authRequiredOperations[opName] {
			middleware.ValidateJWT(srv).ServeHTTP(w, r)
		} else {
			srv.ServeHTTP(w, r)
		}
	}))

	router.Handle("/schema", srv)

	router.Handle("/", playground.Handler("GraphQL Playground", "/query"))

	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{"Authorization", "Content-Type", "X-GraphQL-Operation-Name"}),
	)

	log.Printf("connect to http://localhost:%s/ for GraphQL Playground", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(router)))
}

func mergeOperations(maps ...map[string]bool) map[string]bool {
	result := make(map[string]bool)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func getAdminOperations() map[string]bool {
	return map[string]bool{
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
		"AddTrivia":                      true,
		"DeleteTrivia":                   true,
		"EditTrivia":                     true,
		"GetAdminTrivia":                 true,
		"GetTrivias":                     true,
	}
}

func getCustomerOperations() map[string]bool {
	return map[string]bool{
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
		"AddAnswer":                 true,
		"FilledTemplates":           true,
		"GetFilledTemplates":        true,
		"FilledWouldYouRather":      true,
		"GetWouldYouRathers":        true,
		"GetFilledWouldRathers":     true,
		"TemplatePlayWithFriends":   true,
		"GetFriendTemplates":        true,
		"OwnGame":                   true,
		"GetTrivia":                 true,
		"GetTriviaDetail":           true,
		"GetAnsweredTrivia":         true,
		"GetRequestedTrivia":        true,
		"AcceptRejectTrivia":        true,
		"AddTriviaAnswer":           true,
		"GetAllCategory":            true,
		"DeleteAccount":             true,
	}
}
