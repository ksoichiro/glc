package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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
	ExitCodeError = 1
)

var (
	resolvedUrl   string
	resolvedToken string
	outputPath    string
)

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
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(ExitCodeError)
	}

	switch os.Args[1] {
	case "issues", "i":
	default:
		printUsage()
		os.Exit(ExitCodeError)
	}

	// Options for command
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var (
		token = fs.String("token", "", "Your private token.")
		url   = fs.String("url", "", "GitLab root URL.")
		out   = fs.String("out", "", "Output CSV file.")
	)
	fs.Parse(os.Args[2:])

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	} else {
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
						resolvedToken = v
					case "url":
						resolvedUrl = v
					}
				}
			}
		} else {
			fmt.Println(err)
		}
		defer configFile.Close()
	}

	if *token != "" {
		resolvedToken = *token
	}
	if *url != "" {
		resolvedUrl = *url
	}

	if resolvedToken == "" {
		fmt.Println("Private token is required(-token)")
		return
	}
	if resolvedUrl == "" {
		fmt.Println("GitLab URL is required(-url)")
		return
	}

	if *out == "" {
		fmt.Println("Output file name is required(-out)")
		return
	}
	outputPath = *out

	if len(os.Args) == 1 {
		printUsage()
		os.Exit(ExitCodeError)
	}

	getIssues()
}

func getIssues() {
	var modUrl = resolvedUrl
	if strings.HasSuffix(modUrl, "/") {
		modUrl = strings.TrimSuffix(modUrl, "/")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", modUrl+"/api/v3/issues", nil)
	req.Header.Add("PRIVATE-TOKEN", resolvedToken)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var issues []Issue
	err = json.Unmarshal(body, &issues)
	if err != nil {
		fmt.Println("error: ", err)
	}

	outFile, _ := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE, 0600)
	sjisWriter := transform.NewWriter(outFile, japanese.ShiftJIS.NewEncoder())
	writer := csv.NewWriter(sjisWriter)

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
	fmt.Fprintf(os.Stderr, `glc - GitLab command line interface, especially for managing issues.
Usage: %s command
command:
  issues   get issues
  projects get projects
`, os.Args[0])
}
