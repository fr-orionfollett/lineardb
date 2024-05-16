package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type Creator struct {
	DisplayName string `json:"displayName"`
}

type Assignee struct {
	DisplayName string `json:"displayName"`
}

type NamedField struct {
	Name string `json:"name"`
}

type NumberedField struct {
	Number int `json:"number"`
}

type Labels struct {
	LabelNodes []NamedField `json:"nodes"`
}

type Node struct {
	Id          string        `json:"id"`
	Title       string        `json:"title"`
	CreatedAt   time.Time     `json:"createdAt"`
	CompletedAt time.Time     `json:"completedAt"`
	CanceledAt  time.Time     `json:"canceledAt"`
	Number      int           `json:"number"`
	Estimate    int           `json:"estimate"`
	StartedAt   time.Time     `json:"startedAt"`
	State       NamedField    `json:"state"`
	Project     NamedField    `json:"project"`
	Labels      Labels        `json:"labels"`
	Creator     Creator       `json:"creator"`
	Assignee    Assignee      `json:"assignee"`
	Description string        `json:"description"`
	Url         string        `json:"url"`
	Cycle       NumberedField `json:"cycle"`
}

type Issue struct {
	Nodes    []Node   `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

type Data struct {
	Issues Issue `json:"issues"`
}

type Response struct {
	Data Data `json:"data"`
}

func collectAllIssuesFromLinear() {
	query := `
	query Issues($after: String) {
		issues(first:150, after: $after) {
		  nodes {
			id
			title
			createdAt
			completedAt
			startedAt
			number
			estimate
			canceledAt
			state {
			  name
			}
			labels{
			  nodes{
				name
			  }
			}
			project{
			  name
			}
			creator {
			  displayName
			}
			description
			url
			cycle{
			  number
			}
		  }
		  pageInfo {
			hasNextPage
			endCursor
		  }
		}
	  }`

	if len(os.Args) != 2 {
		log.Fatal("Make sure to include a Linear API key as the first argument!")
	}

	LINEAR_API_KEY := os.Args[1]

	// remove existing db if there is one
	os.Remove("./issues.db")
	db, err := sql.Open("sqlite3", "./issues.db")
	defer db.Close()

	_, err = db.Exec(`create table issues (
		id text primary key,
		title text,
		createdAt text,
		completedAt text,
		startedAt text,
		state text,
		creator text,
		assignee text,
		description text,
		url text,
		canceledAt text,
		number int,
		estimate int,
		labels text,
		project text,
		cycle int
	);`)
	check(err)

	transaction, err := db.Begin()
	insertStmt, err := transaction.Prepare("insert into issues(id, title, createdAt, completedAt, state, creator, description, url, assignee, startedAt, canceledAt, number, estimate, labels, project, cycle) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	check(err)
	defer insertStmt.Close()

	request := map[string]interface{}{
		"query": string(query),
		"variables": map[string]interface{}{
			"after": nil,
		},
	}

	for {
		//for i := 0; i < 5; i++ { // shorter loop for testing purposes

		jsonReq, err := json.Marshal(request)
		check(err)

		client := http.Client{}
		req, err := http.NewRequest("POST", "https://api.linear.app/graphql", bytes.NewBuffer(jsonReq))
		check(err)

		req.Header.Set("Authorization", LINEAR_API_KEY)
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		check(err)
		data, err := io.ReadAll(res.Body)
		check(err)

		var loadedData Response

		err = json.Unmarshal(data, &loadedData)
		check(err)

		// Insert Data
		for _, issue := range loadedData.Data.Issues.Nodes {
			var labels strings.Builder
			for _, i := range issue.Labels.LabelNodes {
				labels.WriteString(i.Name)
				labels.WriteString(", ")
			}
			_, err = insertStmt.Exec(issue.Id, issue.Title, issue.CreatedAt, issue.CompletedAt, issue.State.Name, issue.Creator.DisplayName, issue.Description, issue.Url, issue.Assignee.DisplayName, issue.StartedAt, issue.CanceledAt, issue.Number, issue.Estimate, labels.String(), issue.Project.Name, issue.Cycle.Number)
			if err != nil {
				log.Fatal(err)
			}
		}

		// get cursor
		hasNext := loadedData.Data.Issues.PageInfo.HasNextPage
		endCursor := loadedData.Data.Issues.PageInfo.EndCursor
		fmt.Println("Cursor: " + endCursor)

		if !hasNext {
			break
		}
		request["variables"].(map[string]interface{})["after"] = endCursor
	}
	err = transaction.Commit()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Data loaded successfully!")
}

func main() {

	/*
		This go script gets all issues from the linear api and loads it into a sqlite3 file called "issues.db".
	*/
	fmt.Println("Getting all data from Linear...")

	collectAllIssuesFromLinear()

	fmt.Println("Done!")
}
