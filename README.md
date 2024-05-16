# README

## Description:
This script pulls all Linear issues down into a sqlite3 file. It consists of one go file, main.go.

The sqlite DB gets created in whatever directory you run the command, it will be called "issues.db".
The DB has one table, called "issues".

## Instructions:
In order to run this script, you will need to install go:
```
https://go.dev/doc/install
```

To run the script:
```
    go run https://github.com/FirstResonance/linear-data-wrangling <YOUR LINEAR API KEY HERE> 
```

Or you could clone the repo and run: 
```
    go run main.go <YOUR LINEAR API KEY HERE>
```

To execute adhoc sql queries on the data, you will need a way to interact with sqlite files. I use
sqlite3, which is similar to psql for postgres. 

Here are the docs for sqlite3: https://www.sqlite.org/cli.html

## Schema Description:

The DB has one table called issues, which has this schema:

id 

title

createdAt

completedAt

startedAt

state - In progress, todo, ...

creator

assignee

description

url

canceledAt

number - ION-9999, the number part of that

estimate - complexity estimate (points)

labels - string that is comma separated list of labels

project - project name

cycle - cycle number
