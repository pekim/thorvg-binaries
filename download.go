package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v81/github"
)

const owner = "pekim"
const repo = "thorvg-binaries"

type download struct {
	apiToken   string
	destDir    string
	apiClient  *github.Client
	httpClient *http.Client
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("expected 2 arguments, API token and destination directory")
		os.Exit(1)
	}

	apiToken := os.Args[1]
	destDir := os.Args[2]
	apiClient := github.NewClient(nil).WithAuthToken(apiToken)
	httpClient := &http.Client{Transport: &http.Transport{}}

	dl := download{
		apiToken:   apiToken,
		destDir:    destDir,
		apiClient:  apiClient,
		httpClient: httpClient,
	}
	dl.processArtifacts()
	dl.generateConstantGoFile()
}

func (dl download) processArtifacts() {
	// list artifacts
	listOptions := github.ListArtifactsOptions{}
	artifacts, _, err := dl.apiClient.Actions.ListArtifacts(
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
			dl.processArtifact(artifact)
		} else {
			break
		}
	}
}

func (dl download) processArtifact(artifact *github.Artifact) {
	fmt.Printf("download '%s' from %s\n", *artifact.Name, *artifact.ArchiveDownloadURL)

	req, err := http.NewRequest("GET", *artifact.ArchiveDownloadURL, nil)
	assertNoError(err)
	req.Header.Add("Authorization", "Bearer "+dl.apiToken)

	resp, err := dl.httpClient.Do(req)
	assertNoError(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("response code", resp.Status)
		os.Exit(1)
	}

	content := &bytes.Buffer{}
	_, err = io.Copy(content, resp.Body)
	assertNoError(err)

	reader, err := zip.NewReader(bytes.NewReader(content.Bytes()), int64(content.Len()))
	assertNoError(err)

	for _, file := range reader.File {
		dl.extractFile(file)

		if !strings.HasSuffix(file.Name, ".h") {
			dl.generateLibraryGoFile(file.Name, file.CRC32)
		}
	}
}

func (dl download) extractFile(file *zip.File) {
	fmt.Printf("  extract : %s\n", file.Name)

	destPath := filepath.Join(dl.destDir, file.Name)
	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	assertNoError(err)
	defer destFile.Close()

	srcFile, err := file.Open()
	assertNoError(err)
	defer srcFile.Close()

	_, err = io.Copy(destFile, srcFile)
	assertNoError(err)
}

func (dl download) writeGoFile(filename string, contents string) {
	// write file
	filepath := filepath.Join(dl.destDir, filename) + ".go"
	err := os.WriteFile(filepath, []byte(contents), 0644)
	assertNoError(err)

	// format file
	cmd := exec.Command("gofmt", "-w", filepath)
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println(string(output))
	}
	assertNoError(err)
}

func (dl download) generateLibraryGoFile(filename string, fileCRC uint32) {
	src := fmt.Sprintf(`
// This is a generated file. DO NOT EDIT.

package thorvg

import _ "embed"

//go:embed %s
var sharedObject []byte

const sharedObjectID = "%08x"
`,
		filename, fileCRC,
	)

	dl.writeGoFile(filename, src)
}

func (dl download) generateConstantGoFile() {
	commitHash, err := os.ReadFile("thorvg-commit")
	assertNoError(err)

	src := fmt.Sprintf(`
// This is a generated file. DO NOT EDIT.

package thorvg

const libthorvgCommit = "%s"
`,
		strings.TrimSpace(string(commitHash)),
	)

	dl.writeGoFile("libthorvg-constant.go", src)
}

func assertNoError(err error) {
	if err != nil {
		panic(err)
	}
}
