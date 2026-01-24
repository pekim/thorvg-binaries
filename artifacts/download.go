package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-github/v81/github"
)

const owner = "pekim"
const repo = "thorvg-binaries"

func main() {
	// create clients
	apiClient := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_PAT"))
	httpClient := &http.Client{Transport: &http.Transport{}}

	// list artifacts
	listOptions := github.ListArtifactsOptions{}
	artifacts, _, err := apiClient.Actions.ListArtifacts(
		context.Background(),
		owner, repo,
		&listOptions,
	)
	if err != nil {
		panic(err)
	}

	// Download all of the artifacts from the most recent worklfow run
	//
	// Artifacts are listed in reverse chronological order.
	// So use the first artifacts, until one from a different run is encountered.
	var workflowRunId *int64
	for _, artifact := range artifacts.Artifacts {
		if workflowRunId == nil {
			workflowRunId = artifact.WorkflowRun.ID
		}

		if *artifact.WorkflowRun.ID == *workflowRunId {
			downloadArtifact(httpClient, artifact)
		} else {
			break
		}
	}
}

func downloadArtifact(httpClient *http.Client, artifact *github.Artifact) {
	fmt.Printf("download '%s' from %s\n", *artifact.Name, *artifact.ArchiveDownloadURL)

	req, err := http.NewRequest("GET", *artifact.ArchiveDownloadURL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+os.Getenv("GITHUB_PAT"))

	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("response code", resp.Status)
		os.Exit(1)
	}

	content := &bytes.Buffer{}
	io.Copy(content, resp.Body)

	reader, err := zip.NewReader(bytes.NewReader(content.Bytes()), int64(content.Len()))
	if err != nil {
		panic(err)
	}

	for _, file := range reader.File {
		extractFile(file)
	}
}

func extractFile(file *zip.File) {
	fmt.Printf("  extract : %s\n", file.Name)

	destPath := filepath.Join("artifacts", file.Name)
	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer destFile.Close()

	srcFile, err := file.Open()
	if err != nil {
		panic(err)
	}
	defer srcFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		panic(err)
	}
}
