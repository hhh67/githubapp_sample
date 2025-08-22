package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v60/github"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	appID, err := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse GITHUB_APP_ID: %v", err)
	}

	installationID, err := strconv.ParseInt(os.Getenv("GITHUB_INSTALLATION_ID"), 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse GITHUB_INSTALLATION_ID: %v", err)
	}

	privateKey := os.Getenv("GITHUB_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("GITHUB_PRIVATE_KEY is required")
	}

	transport, err := ghinstallation.New(
		http.DefaultTransport,
		appID,
		installationID,
		[]byte(privateKey),
	)
	if err != nil {
		log.Fatalf("Failed to create GitHub App transport: %v", err)
	}

	httpClient := &http.Client{Transport: transport}
	client := github.NewClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Getting installation access token...")
	token, err := transport.Token(ctx)
	if err != nil {
		log.Fatalf("Failed to get access token: %v", err)
	}
	fmt.Printf("Installation access token: %s\n", token)
	fmt.Printf("Token length: %d characters\n\n", len(token))

	fmt.Println("Fetching repositories...")

	repositories, _, err := client.Apps.ListRepos(ctx, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
		log.Fatalf("Failed to list repositories: %v", err)
	}

	fmt.Printf("Found %d repositories:\n", len(repositories.Repositories))
	for i, repo := range repositories.Repositories {
		fmt.Printf("%d. %s (%s)\n", i+1, repo.GetFullName(), repo.GetHTMLURL())
	}

	fmt.Println("\nAccess token validation...")
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Printf("Warning: Failed to get user info: %v", err)
	} else {
		fmt.Printf("Authenticated as bot: %s\n", user.GetLogin())
	}
}