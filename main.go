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

	if len(repositories.Repositories) > 0 {
		testRepo := repositories.Repositories[0]
		owner := testRepo.GetOwner().GetLogin()
		repoName := testRepo.GetName()
		
		fmt.Printf("\n=== API Testing with %s/%s ===\n", owner, repoName)
		
		fmt.Println("\n1. Testing Repositories.GetContents...")
		testGetContents(ctx, client, owner, repoName)
		
		fmt.Println("\n2. Testing Repositories.DownloadContents...")
		testDownloadContents(ctx, client, owner, repoName)
		
		fmt.Println("\n3. Testing Repositories.ListCommits...")
		testListCommits(ctx, client, owner, repoName)
		
		fmt.Println("\n4. Testing Client.RateLimits...")
		testRateLimits(ctx, client)
	} else {
		fmt.Println("\nNo repositories available for API testing")
	}
}

func testGetContents(ctx context.Context, client *github.Client, owner, repo string) {
	content, _, resp, err := client.Repositories.GetContents(ctx, owner, repo, "README.md", nil)
	if err != nil {
		fmt.Printf("❌ GetContents failed: %v\n", err)
		return
	}
	fmt.Printf("✅ GetContents success (Status: %d)\n", resp.StatusCode)
	if content != nil {
		fmt.Printf("   File: %s, Size: %d bytes\n", content.GetName(), content.GetSize())
	}
}

func testDownloadContents(ctx context.Context, client *github.Client, owner, repo string) {
	_, resp, err := client.Repositories.DownloadContents(ctx, owner, repo, "README.md", nil)
	if err != nil {
		fmt.Printf("❌ DownloadContents failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("✅ DownloadContents success (Status: %d)\n", resp.StatusCode)
	fmt.Printf("   Content-Type: %s\n", resp.Header.Get("Content-Type"))
}

func testListCommits(ctx context.Context, client *github.Client, owner, repo string) {
	commits, resp, err := client.Repositories.ListCommits(ctx, owner, repo, &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 5},
	})
	if err != nil {
		fmt.Printf("❌ ListCommits failed: %v\n", err)
		return
	}
	fmt.Printf("✅ ListCommits success (Status: %d)\n", resp.StatusCode)
	fmt.Printf("   Found %d commits\n", len(commits))
	if len(commits) > 0 {
		latest := commits[0]
		fmt.Printf("   Latest: %s by %s\n", 
			latest.GetSHA()[:7], 
			latest.GetCommit().GetAuthor().GetName())
	}
}

func testRateLimits(ctx context.Context, client *github.Client) {
	rateLimits, resp, err := client.RateLimits(ctx)
	if err != nil {
		fmt.Printf("❌ RateLimits failed: %v\n", err)
		return
	}
	fmt.Printf("✅ RateLimits success (Status: %d)\n", resp.StatusCode)
	if rateLimits != nil && rateLimits.Core != nil {
		fmt.Printf("   Core API: %d/%d remaining (Reset: %s)\n",
			rateLimits.Core.Remaining,
			rateLimits.Core.Limit,
			rateLimits.Core.Reset.Format("15:04:05"))
	}
}