package main

import (
	"context"
	"fmt"
	"log"

	"github.com/igun997/isometricicon-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	apiClient := client.New()

	s := server.NewMCPServer(
		"isometricon",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Login tool
	loginTool := mcp.NewTool("login",
		mcp.WithDescription("Login to IsometricIcon with your email and password"),
		mcp.WithString("email",
			mcp.Required(),
			mcp.Description("Your IsometricIcon account email"),
		),
		mcp.WithString("password",
			mcp.Required(),
			mcp.Description("Your IsometricIcon account password"),
		),
	)
	s.AddTool(loginTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		email, err := req.RequireString("email")
		if err != nil {
			return mcp.NewToolResultError("email is required"), nil
		}
		password, err := req.RequireString("password")
		if err != nil {
			return mcp.NewToolResultError("password is required"), nil
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
