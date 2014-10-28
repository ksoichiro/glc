package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"code.google.com/p/go.text/encoding/japanese"
	"code.google.com/p/go.text/transform"
)

const (
	ExitCodeError    = 1
	EncodingShiftJIS = "sjis"
	EncodingUTF8     = "utf8"
)

var (
	ResolvedUrl   string
	ResolvedToken string
	OutputPath    string
	CsvEncoding   string
	ProjectId     string
)

type Project struct {
	Id                int
	Description       string
	Public            bool
	VisibilityLevel   int    `json:"visibility_level"`
	SshUrlToRepo      string `json:"ssh_url_to_repo"`
	HttpUrlToRepo     string `json:"http_url_to_repo"`
	WebUrl            string
	Name              string
	NameWithNamespace string `json:"name_with_namespace"`
	Path              string
	PathWithNamespace string `json:"path_with_namespace"`
	IssuesEnabled     bool
	CreatedAt         string `json:"created_at"`
}

type Issue struct {
	Id          int
	Iid         int
	ProjectId   int `json:"project_id"`
	Title       string
	Description string
	Assignee    User
	Author      User
	State       string
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type User struct {
	Id        int
	Username  string
	Email     string
	Name      string
	State     string
	CreatedAt string `json:"created_at"`
}

func main() {
	// Read user's config if exists
	usr, err := user.Current()
	if err != nil {
		os.Exit(ExitCodeError)
	}
	configPath := filepath.Join(usr.HomeDir, ".glc")
	configFile, err := os.OpenFile(configPath, os.O_RDONLY, 0600)
	if err == nil || os.IsExist(err) {
		lines := []string{}
		scanner := bufio.NewScanner(configFile)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if serr := scanner.Err(); serr != nil {
			fmt.Fprintf(os.Stderr, "Failed to scan file %s: %v\n", configPath, err)
		} else {
			for _, line := range lines {
				r := regexp.MustCompile("^[ \t]*([^=]+)[ \t]*=[ \t]*(.*)$")
				groups := r.FindStringSubmatch(line)
				if groups == nil {
					continue
				}
				k := groups[1]
				v := groups[2]
				switch k {
				case "token":
					ResolvedToken = v
				case "url":
					ResolvedUrl = v
				}
			}
		}
	}
	defer configFile.Close()

	// Global options
	var (
		token       = flag.String("token", "", "Your private token.")
		url         = flag.String("url", "", "GitLab root URL.")
		out         = flag.String("out", "", "Output CSV file.")
		csvEncoding = flag.String("csvEncoding", EncodingShiftJIS, "Output encoding for CSV file.")
	)
	flag.Parse()

	// Resolve global settings.
	// Command line options override user's config
	if *token != "" {
		ResolvedToken = *token
	}
	if ResolvedToken == "" {
		fmt.Fprintln(os.Stderr, "Private token is required(-token)\n")
		printUsage()
		os.Exit(ExitCodeError)
	}
	if *url != "" {
		ResolvedUrl = *url
	}
	if ResolvedUrl == "" {
		fmt.Fprintln(os.Stderr, "GitLab URL is required(-url)\n")
		printUsage()
		os.Exit(ExitCodeError)
	}
	if *out == "" {
		fmt.Fprintln(os.Stderr, "Output file name is required(-out)\n")
		printUsage()
		os.Exit(ExitCodeError)
	}
	OutputPath = *out

	if *csvEncoding == "" || (*csvEncoding != EncodingUTF8 && *csvEncoding != EncodingShiftJIS) {
		*csvEncoding = EncodingShiftJIS
	}
	CsvEncoding = *csvEncoding

	// Get objective: projects, issues, ...
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Command is required.\n")
		printUsage()
		os.Exit(ExitCodeError)
	}

	switch flag.Arg(0) {
	case "projects", "p":
		getProjects()
	case "issues", "i":
		getIssues()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", flag.Arg(0))
		printUsage()
		os.Exit(ExitCodeError)
	}
}

func getProjects() {
	var modUrl = ResolvedUrl
	if strings.HasSuffix(modUrl, "/") {
		modUrl = strings.TrimSuffix(modUrl, "/")
	}

	client := &http.Client{}
	path := "/api/v3/projects"
	req, err := http.NewRequest("GET", modUrl+path, nil)
	req.Header.Add("PRIVATE-TOKEN", ResolvedToken)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error while executing request: %v\n\n", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var projects []Project
	err = json.Unmarshal(body, &projects)
	if err != nil {
		fmt.Println("error while unmarshaling json: %v\n\n", err)
		fmt.Println(string(body))
	}

	outFile, _ := os.OpenFile(OutputPath, os.O_WRONLY|os.O_CREATE, 0600)
	var writer *csv.Writer
	if CsvEncoding == EncodingShiftJIS {
		sjisWriter := transform.NewWriter(outFile, japanese.ShiftJIS.NewEncoder())
		writer = csv.NewWriter(sjisWriter)
	} else {
		writer = csv.NewWriter(outFile)
	}

	writer.Write([]string{"Id", "Name", "NameWithNamespace", "Path", "PathWithNamespace", "IssuesEnabled", "CreatedAt"})
	for _, project := range projects {
		writer.Write([]string{
			fmt.Sprintf("%d", project.Id),
			project.Name,
			project.NameWithNamespace,
			project.Path,
			project.PathWithNamespace,
			fmt.Sprintf("%t", project.IssuesEnabled),
			project.CreatedAt,
		})
	}
	writer.Flush()
}

func getIssues() {
	// Get command options
	var projectId = flag.String("project", "", "Target project ID. Optional.")
	flag.Parse()
	if *projectId != "" {
		ProjectId = *projectId
	}

	var modUrl = ResolvedUrl
	if strings.HasSuffix(modUrl, "/") {
		modUrl = strings.TrimSuffix(modUrl, "/")
	}

	client := &http.Client{}
	path := "/api/v3"
	if ProjectId != "" {
		path += "/projects/" + strings.Replace(ProjectId, "/", "%2F", 1)
	}
	path += "/issues"
	req, err := http.NewRequest("GET", modUrl+path, nil)
	req.Header.Add("PRIVATE-TOKEN", ResolvedToken)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error while executing request: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var issues []Issue
	err = json.Unmarshal(body, &issues)
	if err != nil {
		fmt.Println("error while unmarshaling json: ", err)
		fmt.Println(string(body))
	}

	outFile, _ := os.OpenFile(OutputPath, os.O_WRONLY|os.O_CREATE, 0600)
	var writer *csv.Writer
	if CsvEncoding == EncodingShiftJIS {
		sjisWriter := transform.NewWriter(outFile, japanese.ShiftJIS.NewEncoder())
		writer = csv.NewWriter(sjisWriter)
	} else {
		writer = csv.NewWriter(outFile)
	}

	writer.Write([]string{"Id", "ProjectId", "Title", "Descrption", "Assignee", "Author", "State", "UpdatedAt", "CreatedAt"})
	for _, issue := range issues {
		writer.Write([]string{
			fmt.Sprintf("%d", issue.Id),
			fmt.Sprintf("%d", issue.ProjectId),
			issue.Title,
			issue.Description,
			issue.Assignee.Name,
			issue.Author.Name,
			issue.State,
			issue.UpdatedAt,
			issue.CreatedAt,
		})
	}
	writer.Flush()
}

func printUsage() {
	fmt.Fprintln(os.Stderr,
		`usage: glc [-token=<private_token>] [-url=<gitlab_url>]
           -out=<csv_file_path> [-csvEncoding=<encoding>]
           <command> [<args>]
options:
  -token        Your private token.
                This is required unless you define it in ~/.glc.
  -url          GitLab root URL.
                This is required unless you define it in ~/.glc.
  -out          Output CSV file. Required.
  -csvEncoding  Output encoding for CSV file.
                "sjis" and "utf8" is available.
                "sjis" is default.
command:
  p[rojects]    Get/update projects
  i[ssues]      Get/update issues
`)
}
