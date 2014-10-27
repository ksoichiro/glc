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
	ExitCodeError    = 1
	EncodingShiftJIS = "sjis"
	EncodingUTF8     = "utf8"
)

var (
	ResolvedUrl   string
	ResolvedToken string
	OutputPath    string
	CsvEncoding   string
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
		token       = fs.String("token", "", "Your private token.")
		url         = fs.String("url", "", "GitLab root URL.")
		out         = fs.String("out", "", "Output CSV file.")
		csvEncoding = fs.String("csvEncoding", EncodingShiftJIS, "Output encoding for CSV file.")
	)
	fs.Parse(os.Args[2:])

	if *csvEncoding == "" || (*csvEncoding != EncodingUTF8 && *csvEncoding != EncodingShiftJIS) {
		*csvEncoding = EncodingShiftJIS
	}
	CsvEncoding = *csvEncoding

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
						ResolvedToken = v
					case "url":
						ResolvedUrl = v
					}
				}
			}
		} else {
			fmt.Println(err)
		}
		defer configFile.Close()
	}

	if *token != "" {
		ResolvedToken = *token
	}
	if *url != "" {
		ResolvedUrl = *url
	}

	if ResolvedToken == "" {
		fmt.Println("Private token is required(-token)")
		return
	}
	if ResolvedUrl == "" {
		fmt.Println("GitLab URL is required(-url)")
		return
	}

	if *out == "" {
		fmt.Println("Output file name is required(-out)")
		return
	}
	OutputPath = *out

	if len(os.Args) == 1 {
		printUsage()
		os.Exit(ExitCodeError)
	}

	getIssues()
}

func getIssues() {
	var modUrl = ResolvedUrl
	if strings.HasSuffix(modUrl, "/") {
		modUrl = strings.TrimSuffix(modUrl, "/")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", modUrl+"/api/v3/issues", nil)
	req.Header.Add("PRIVATE-TOKEN", ResolvedToken)
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
	fmt.Fprintf(os.Stderr, `glc - GitLab command line interface, especially for managing issues.
Usage: %s command
command:
  issues   get issues
  projects get projects
`, os.Args[0])
}
