package workflow

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"

	"workflow-code-test/api/internal/repository"
	"workflow-code-test/api/internal/service"
)

type Service struct {
	db              *pgx.Conn
	workflowService *service.WorkflowService
}

func NewService(db *pgx.Conn) (*Service, error) {
	// Create repository
	workflowRepo := repository.NewWorkflowRepository(db)

	// Create service
	workflowService := service.NewWorkflowService(workflowRepo)

	return &Service{
		db:              db,
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
