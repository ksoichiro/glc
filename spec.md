# Spec consideration

## Available API

Although we can use CRUD APIs, but they should be used mainly for reference.
To update projects, issues, etc. should be executed with Git operation or GitLab GUI.

## Command format design

It should be simple, short and flexible.

```sh
glc i g[et] [-p[roject]=id] [-status=open|closed]
glc i n[ew] -p[roject]=id -t[itle]=title -d[escription]=description
glc i update status close -p[roject]=id -i[d]=id
glc i r[eopen] -p[roject]=id -i[d]=id
```

## Printing

Printing GitLab data in several way is one of the big purposes of developing this tool.

### Candidates:

#### CSV

This is simple and we can handle it easily by using `encoding/csv` package.

#### Excel

For managing and analyzing project status, using Excel format is important.
It is better to export data directly to Excel file than to CSV file.

#### JSON/XML

Exporting data with JSON or XML is very easy in Golang,
and it's useful when we want to use data in other programs.

## Problems

### Filtering data

Max records is limited by default.

### Joining with other tables

For example, to get issues of a specific project, we must set ID of the project.
But we don't know it usually.
To do that, at first we should use `/projects` API and then use `/projects/:id/issues` API.

