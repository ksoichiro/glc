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
	var (
		token = flag.String("token", "", "Your private token.")
		url   = flag.String("url", "", "GitLab root URL.")
		out   = flag.String("out", "", "Output CSV file.")
	)
	flag.Parse()

	var resolvedToken string
	var resolvedUrl string
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

	outFile, _ := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE, 0600)
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
