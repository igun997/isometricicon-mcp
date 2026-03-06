package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/igun997/isometricicon-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	apiClient := client.New()

	// Auto-login from env vars if not already logged in from stored token
	if !apiClient.IsLoggedIn() {
		email := os.Getenv("ISOMETRICON_EMAIL")
		password := os.Getenv("ISOMETRICON_PASSWORD")
		if email != "" && password != "" {
			apiClient.Login(email, password)
		}
	}

	s := server.NewMCPServer(
		"isometricon",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Login tool
	loginTool := mcp.NewTool("login",
		mcp.WithDescription("Login to IsometricIcon. Uses ISOMETRICON_EMAIL/ISOMETRICON_PASSWORD env vars if parameters are omitted."),
		mcp.WithString("email",
			mcp.Description("Your IsometricIcon account email (optional if ISOMETRICON_EMAIL is set)"),
		),
		mcp.WithString("password",
			mcp.Description("Your IsometricIcon account password (optional if ISOMETRICON_PASSWORD is set)"),
		),
	)
	s.AddTool(loginTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		email, _ := req.RequireString("email")
		password, _ := req.RequireString("password")

		if email == "" {
			email = os.Getenv("ISOMETRICON_EMAIL")
		}
		if password == "" {
			password = os.Getenv("ISOMETRICON_PASSWORD")
		}

		if email == "" || password == "" {
			return mcp.NewToolResultError("email and password are required — provide as parameters or set ISOMETRICON_EMAIL and ISOMETRICON_PASSWORD env vars"), nil
		}

		name, err := apiClient.Login(email, password)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Login failed: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Logged in as %s", name)), nil
	})

	// Generate icon tool
	generateTool := mcp.NewTool("generate_icon",
		mcp.WithDescription("Generate an isometric icon from a text prompt. Requires login first."),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Text description of the icon to generate"),
		),
		mcp.WithString("output_path",
			mcp.Description("File path to save the PNG image (default: ./icon.png)"),
		),
	)
	s.AddTool(generateTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prompt, err := req.RequireString("prompt")
		if err != nil {
			return mcp.NewToolResultError("prompt is required"), nil
		}

		outputPath, _ := req.RequireString("output_path")

		result, err := apiClient.GenerateIcon(prompt, outputPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Generation failed: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Icon generated!\nCDN URL: %s\nSaved to: %s", result.URL, result.FilePath)), nil
	})

	// Check credits tool
	creditsTool := mcp.NewTool("check_credits",
		mcp.WithDescription("Check your IsometricIcon credit balance. Requires login first."),
	)
	s.AddTool(creditsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		credits, err := apiClient.CheckCredits()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Credits check failed: %v", err)), nil
		}

		unlimited := "No"
		if credits.Unlimited {
			unlimited = "Yes"
		}

		return mcp.NewToolResultText(fmt.Sprintf("Credits: %d\nUnlimited: %s", credits.Balance, unlimited)), nil
	})

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
