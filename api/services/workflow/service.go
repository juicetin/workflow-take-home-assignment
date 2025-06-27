package workflow

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"

	"workflow-code-test/api/internal/repository"
	"workflow-code-test/api/internal/service"
	"workflow-code-test/api/pkg/db"
)

type Service struct {
	db              *pgx.Conn
	sqlDB           *sql.DB
	workflowService *service.WorkflowService
}

func NewService(conn *pgx.Conn, config *db.Config) (*Service, error) {
	// Create sql.DB connection for Jet repository
	sqlDB, err := db.GetJetDB(config)
	if err != nil {
		return nil, err
	}

	// Create repository using sql.DB
	workflowRepo := repository.NewWorkflowRepository(sqlDB)

	// Create service
	workflowService := service.NewWorkflowService(workflowRepo)

	return &Service{
		db:              conn,
		sqlDB:           sqlDB,
		workflowService: workflowService,
	}, nil
}

// jsonMiddleware sets the Content-Type header to application/json
func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (s *Service) LoadRoutes(parentRouter *mux.Router, isProduction bool) {
	router := parentRouter.PathPrefix("/workflows").Subrouter()
	router.StrictSlash(false)
	router.Use(jsonMiddleware)

	router.HandleFunc("/{id}", s.HandleGetWorkflow).Methods("GET")
	router.HandleFunc("/{id}/execute", s.HandleExecuteWorkflow).Methods("POST")

}
