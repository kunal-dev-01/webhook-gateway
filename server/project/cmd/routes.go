package main

import (
	"github.com/gofiber/fiber/v2"
	"shared/middleware"
)

func SetupRoutes(app *fiber.App, deps *Dependencies) {
	authMiddleware := middleware.NewAuthMiddleware(deps.JWTManager, deps.Redis.Client)
	rateLimiter := middleware.NewRateLimiter(deps.Redis.Client)

	v1 := app.Group("/api/v1")

	protected := v1.Group("")
	protected.Use(authMiddleware.RequireAuth())
	protected.Use(rateLimiter.ByUser(middleware.GeneralRateLimit()))

	// ── Projects ─────────────────────────────────────────────────────────────
	//
	// POST   /user/organizations/:org_id/projects                          → Create project
	// GET    /user/organizations/:org_id/projects                          → List projects
	// GET    /user/organizations/:org_id/projects/:project_id              → Get project
	// PATCH  /user/organizations/:org_id/projects/:project_id              → Update project
	// DELETE /user/organizations/:org_id/projects/:project_id              → Delete project
	// POST   /user/organizations/:org_id/projects/:project_id/toggle       → Toggle active
	//
	// ── Endpoints ────────────────────────────────────────────────────────────
	//
	// POST   /user/organizations/:org_id/projects/:project_id/endpoints                        → Create endpoint
	// GET    /user/organizations/:org_id/projects/:project_id/endpoints                        → List endpoints
	// GET    /user/organizations/:org_id/projects/:project_id/endpoints/:endpoint_id           → Get endpoint
	// PATCH  /user/organizations/:org_id/projects/:project_id/endpoints/:endpoint_id           → Update endpoint
	// DELETE /user/organizations/:org_id/projects/:project_id/endpoints/:endpoint_id           → Delete endpoint
	// POST   /user/organizations/:org_id/projects/:project_id/endpoints/:endpoint_id/toggle    → Toggle active
	//
	// ── Custom Domains ───────────────────────────────────────────────────────
	//
	// POST   /user/organizations/:org_id/domains                           → Add domain
	// GET    /user/organizations/:org_id/domains                           → List domains
	// GET    /user/organizations/:org_id/domains/:domain_id                → Get domain + CNAME instructions
	// DELETE /user/organizations/:org_id/domains/:domain_id                → Remove domain
	// POST   /user/organizations/:org_id/domains/:domain_id/verify         → Trigger DNS check

	orgs := protected.Group("/user/organizations/:org_id")

	// Projects
	projects := orgs.Group("/projects")
	projects.Post("/", deps.ProjectHandler.CreateProject)
	projects.Get("/", deps.ProjectHandler.ListProjects)
	projects.Get("/:project_id", deps.ProjectHandler.GetProject)
	projects.Patch("/:project_id", deps.ProjectHandler.UpdateProject)
	projects.Delete("/:project_id", deps.ProjectHandler.DeleteProject)
	projects.Post("/:project_id/toggle", deps.ProjectHandler.ToggleProject)

	// Endpoints (nested under project)
	endpoints := projects.Group("/:project_id/endpoints")
	endpoints.Post("/", deps.EndpointHandler.CreateEndpoint)
	endpoints.Get("/", deps.EndpointHandler.ListEndpoints)
	endpoints.Get("/:endpoint_id", deps.EndpointHandler.GetEndpoint)
	endpoints.Patch("/:endpoint_id", deps.EndpointHandler.UpdateEndpoint)
	endpoints.Delete("/:endpoint_id", deps.EndpointHandler.DeleteEndpoint)
	endpoints.Post("/:endpoint_id/toggle", deps.EndpointHandler.ToggleEndpoint)

	// Custom Domains
	domains := orgs.Group("/domains")
	domains.Post("/", deps.DomainHandler.AddDomain)
	domains.Get("/", deps.DomainHandler.ListDomains)
	domains.Get("/:domain_id", deps.DomainHandler.GetDomain)
	domains.Delete("/:domain_id", deps.DomainHandler.DeleteDomain)
	domains.Post("/:domain_id/verify", deps.DomainHandler.TriggerVerification)

	// API Keys
	apiKeys := orgs.Group("/api-keys")
	apiKeys.Post("/", deps.APIKeyHandler.CreateAPIKey)
	apiKeys.Get("/", deps.APIKeyHandler.ListAPIKeys)
	apiKeys.Get("/:key_id", deps.APIKeyHandler.GetAPIKey)
	apiKeys.Post("/:key_id/revoke", deps.APIKeyHandler.RevokeAPIKey)
	apiKeys.Delete("/:key_id", deps.APIKeyHandler.DeleteAPIKey)

	// CLI Sessions (user-level, not org-scoped)
	// /revoke-all must be registered BEFORE /:session_id to avoid param conflict
	cliSessions := protected.Group("/user/cli-sessions")
	cliSessions.Get("/", deps.CLISessionHandler.ListSessions)
	cliSessions.Post("/revoke-all", deps.CLISessionHandler.RevokeAllSessions)
	cliSessions.Get("/:session_id", deps.CLISessionHandler.GetSession)
	cliSessions.Post("/:session_id/revoke", deps.CLISessionHandler.RevokeSession)
}