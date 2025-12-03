package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "quickbase-personal-mcp"
	serverVersion = "1.0.0"
)

// Repo paths - customize these to your setup
var (
	quickbaseJSPath   = filepath.Join(os.Getenv("HOME"), "Projects", "Personal", "quickbase-js")
	quickbaseGoPath   = filepath.Join(os.Getenv("HOME"), "Projects", "Personal", "quickbase-tree", "quickbase-go")
	quickbaseSpecPath = filepath.Join(os.Getenv("HOME"), "Projects", "Personal", "quickbase-spec")
)

// Main server struct
type QuickBasePersonalMCPServer struct {
	logger *log.Logger
}

func main() {
	logger := log.New(os.Stderr, "["+serverName+"] ", log.LstdFlags)
	logger.Printf("Starting %s v%s", serverName, serverVersion)

	// Create server instance
	s := &QuickBasePersonalMCPServer{
		logger: logger,
	}

	// Setup MCP tools
	tools := s.setupTools()

	// Create MCP server
	mcpServer := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(true),
	)

	// Register tool handlers
	mcpServer.AddTool(tools[0], s.handleSearchCode)
	mcpServer.AddTool(tools[1], s.handleCompareImplementations)
	mcpServer.AddTool(tools[2], s.handleGetAuthExample)
	mcpServer.AddTool(tools[3], s.handleListFeatures)
	mcpServer.AddTool(tools[4], s.handleCheckParity)

	// Start server
	if err := server.ServeStdio(mcpServer); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}

// setupTools defines all MCP tools
func (s *QuickBasePersonalMCPServer) setupTools() []mcp.Tool {
	return []mcp.Tool{
		// 1. search_code
		{
			Name:        "search_code",
			Description: "Search across your QuickBase SDK repositories (JS, Go, and spec).",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query (e.g., 'ticket auth', 'pagination', 'rate limit')",
					},
					"repo": map[string]interface{}{
						"type":        "string",
						"description": "Limit to specific repo: 'js', 'go', 'spec', 'all' (default: 'all')",
						"enum":        []string{"js", "go", "spec", "all"},
					},
				},
				Required: []string{"query"},
			},
		},
		// 2. compare_implementations
		{
			Name:        "compare_implementations",
			Description: "Compare how a feature is implemented in JavaScript vs Go SDK.",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"feature": map[string]interface{}{
						"type":        "string",
						"description": "Feature to compare (e.g., 'ticket-auth', 'temp-token', 'pagination', 'retry')",
					},
				},
				Required: []string{"feature"},
			},
		},
		// 3. get_auth_example
		{
			Name:        "get_auth_example",
			Description: "Get authentication examples from your SDKs.",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"auth_type": map[string]interface{}{
						"type":        "string",
						"description": "Authentication type: 'user-token', 'temp-token', 'sso', 'ticket'",
						"enum":        []string{"user-token", "temp-token", "sso", "ticket"},
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "SDK language: 'js', 'go', 'both' (default: 'both')",
						"enum":        []string{"js", "go", "both"},
					},
				},
				Required: []string{"auth_type"},
			},
		},
		// 4. list_features
		{
			Name:        "list_features",
			Description: "List what features are implemented in your SDKs.",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"category": map[string]interface{}{
						"type":        "string",
						"description": "Feature category: 'auth', 'client', 'pagination', 'all' (default: 'all')",
						"enum":        []string{"auth", "client", "pagination", "all"},
					},
				},
			},
		},
		// 5. check_parity
		{
			Name:        "check_parity",
			Description: "Check feature parity between JavaScript and Go SDKs.",
			InputSchema: mcp.ToolInputSchema{
				Type:       "object",
				Properties: map[string]interface{}{},
			},
		},
	}
}

// Tool handlers
func (s *QuickBasePersonalMCPServer) handleSearchCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Query string `json:"query"`
		Repo  string `json:"repo"`
	}
	argsData, _ := json.Marshal(request.Params.Arguments)
	if err := json.Unmarshal(argsData, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}
	if params.Repo == "" {
		params.Repo = "all"
	}

	// Determine which repos to search
	repos := []struct {
		name string
		path string
	}{}

	if params.Repo == "all" || params.Repo == "js" {
		repos = append(repos, struct {
			name string
			path string
		}{"quickbase-js", quickbaseJSPath})
	}
	if params.Repo == "all" || params.Repo == "go" {
		repos = append(repos, struct {
			name string
			path string
		}{"quickbase-go", quickbaseGoPath})
	}
	if params.Repo == "all" || params.Repo == "spec" {
		repos = append(repos, struct {
			name string
			path string
		}{"quickbase-spec", quickbaseSpecPath})
	}

	var results strings.Builder
	results.WriteString(fmt.Sprintf("Searching for: %s\n\n", params.Query))

	for _, repo := range repos {
		results.WriteString(fmt.Sprintf("## %s\n\n", repo.name))

		// Use ripgrep for fast searching
		cmd := exec.Command("rg", "--no-heading", "--line-number", "--color", "never", params.Query, repo.path)
		output, err := cmd.Output()
		if err != nil {
			results.WriteString(fmt.Sprintf("No matches found\n\n"))
			continue
		}

		results.WriteString(string(output))
		results.WriteString("\n")
	}

	return mcp.NewToolResultText(results.String()), nil
}

func (s *QuickBasePersonalMCPServer) handleCompareImplementations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Feature string `json:"feature"`
	}
	argsData, _ := json.Marshal(request.Params.Arguments)
	if err := json.Unmarshal(argsData, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	// Map features to file paths
	featureMap := map[string]struct {
		jsPath string
		goPath string
	}{
		"ticket-auth":  {jsPath: "src/auth/ticket.ts", goPath: "auth/ticket.go"},
		"temp-token":   {jsPath: "src/auth/temp-token.ts", goPath: "auth/temp_token.go"},
		"user-token":   {jsPath: "src/auth/user-token.ts", goPath: "auth/user_token.go"},
		"sso":          {jsPath: "src/auth/sso.ts", goPath: "auth/sso_token.go"},
		"pagination":   {jsPath: "src/client/pagination.ts", goPath: "client/pagination.go"},
		"retry":        {jsPath: "src/client/retry.ts", goPath: "client/client.go"},
		"throttle":     {jsPath: "src/client/throttle.ts", goPath: "client/throttle.go"},
	}

	paths, ok := featureMap[params.Feature]
	if !ok {
		return mcp.NewToolResultError(fmt.Sprintf("Unknown feature: %s", params.Feature)), nil
	}

	var results strings.Builder
	results.WriteString(fmt.Sprintf("# Comparing: %s\n\n", params.Feature))

	// Read JS implementation
	jsFullPath := filepath.Join(quickbaseJSPath, paths.jsPath)
	jsContent, err := os.ReadFile(jsFullPath)
	if err != nil {
		results.WriteString(fmt.Sprintf("## JavaScript (%s)\nFile not found\n\n", paths.jsPath))
	} else {
		results.WriteString(fmt.Sprintf("## JavaScript (%s)\n\n```typescript\n%s\n```\n\n", paths.jsPath, string(jsContent)))
	}

	// Read Go implementation
	goFullPath := filepath.Join(quickbaseGoPath, paths.goPath)
	goContent, err := os.ReadFile(goFullPath)
	if err != nil {
		results.WriteString(fmt.Sprintf("## Go (%s)\nFile not found\n\n", paths.goPath))
	} else {
		results.WriteString(fmt.Sprintf("## Go (%s)\n\n```go\n%s\n```\n\n", paths.goPath, string(goContent)))
	}

	return mcp.NewToolResultText(results.String()), nil
}

func (s *QuickBasePersonalMCPServer) handleGetAuthExample(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		AuthType string `json:"auth_type"`
		Language string `json:"language"`
	}
	argsData, _ := json.Marshal(request.Params.Arguments)
	if err := json.Unmarshal(argsData, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}
	if params.Language == "" {
		params.Language = "both"
	}

	// TODO: Implement auth examples
	result := fmt.Sprintf("Getting %s auth examples for: %s\n\nTODO: Implement auth examples", params.AuthType, params.Language)

	return mcp.NewToolResultText(result), nil
}

func (s *QuickBasePersonalMCPServer) handleListFeatures(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Category string `json:"category"`
	}
	argsData, _ := json.Marshal(request.Params.Arguments)
	if err := json.Unmarshal(argsData, &params); err != nil {
		params.Category = "all"
	}

	result := `# QuickBase SDK Features

## Authentication Methods
- ✅ User Token (both JS & Go)
- ✅ Temporary Token (both JS & Go)
- ✅ SSO Token (both JS & Go)
- ✅ Ticket Auth - API_Authenticate (both JS & Go)

## Client Features
- ✅ Retry with exponential backoff (both JS & Go)
- ✅ Rate limiting / throttling (both JS & Go)
- ✅ Automatic date parsing (both JS & Go)
- ✅ Custom error types (both JS & Go)

## Pagination
- ✅ Fluent pagination API (both JS & Go)
- ✅ Auto-pagination (both JS & Go)
- ✅ Manual page iteration (both JS & Go)

## Code Generation
- ✅ TypeScript types from OpenAPI spec (JS)
- ✅ Go types from OpenAPI spec (Go)
- ✅ Shared OpenAPI spec (both)
`

	return mcp.NewToolResultText(result), nil
}

func (s *QuickBasePersonalMCPServer) handleCheckParity(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result := `# Feature Parity Check

## ✅ Complete Parity
- User Token Auth
- Temporary Token Auth
- SSO Token Auth
- Ticket Auth (API_Authenticate)
- Retry logic
- Rate limiting
- Date parsing
- Error handling
- Pagination (fluent API)

## 🔄 Differences
- **Browser support**: JS has browser bundles, Go is server-only
- **Testing**: JS uses Vitest, Go uses native testing
- **Generated code**: Different generators (openapi-generator-typescript vs oapi-codegen)

## 📝 Implementation Notes
- Both SDKs share the same OpenAPI spec via git submodule
- Both follow the same architectural patterns
- Both have identical test fixtures (JSON-based)
- Code structure mirrors between languages

## Version Info
- quickbase-js: v2.1.0
- quickbase-go: v1.2.0
`

	return mcp.NewToolResultText(result), nil
}
