package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/pongsathonn/ihavefood/api-gateway/genproto"
)

const ihavefood = `
=======================================================================
██╗██╗  ██╗ █████╗ ██╗   ██╗███████╗███████╗ ██████╗  ██████╗ ██████╗
██║██║  ██║██╔══██╗██║   ██║██╔════╝██╔════╝██╔═══██╗██╔═══██╗██╔══██╗
██║███████║███████║██║   ██║█████╗  █████╗  ██║   ██║██║   ██║██║  ██║
██║██╔══██║██╔══██║╚██╗ ██╔╝██╔══╝  ██╔══╝  ██║   ██║██║   ██║██║  ██║
██║██║  ██║██║  ██║ ╚████╔╝ ███████╗██║     ╚██████╔╝╚██████╔╝██████╔╝
╚═╝╚═╝  ╚═╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚═╝      ╚═════╝  ╚═════╝ ╚═════╝
=======================================================================
`

func cors(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//    For production
		// 	  swg := fmt.Sprintf("http://localhost:%s", os.Getenv("SWAGGER_UI_PORT"))
		// 	  w.Header().Add("Access-Control-Allow-Origin", swg)

		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func prettierJSON(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Accept", "application/json+pretty")
		h.ServeHTTP(w, r)
	})
}

func Run(h http.Handler) error {
	fmt.Print(ihavefood)

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(os.Getenv("AUTH_URI"), opt)
	if err != nil {
		return err
	}
	auth := NewAuthMiddleware(pb.NewAuthServiceClient(conn))

	// FIXME: remove this ?
	// Update role and DELETE methods requires "admin" role.(Just for now)
	http.Handle("PATCH /auth/users/roles", auth.Authz(h))
	http.Handle("DELETE /api/*", auth.Authz(h))
	http.Handle("/api/*", auth.Authn(h))
	http.Handle("/", h)

	slog.Info(fmt.Sprintf("Gateway listening on port :%s", os.Getenv("SERVER_PORT")))
	s := fmt.Sprintf(":%s", os.Getenv("SERVER_PORT"))
	if err := http.ListenAndServe(s, prettierJSON(cors(h))); err != nil {
		return err
	}

	return nil
}
