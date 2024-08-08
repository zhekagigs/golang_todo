# golang_todo

Microbrewery task managment system.

App should be run with one argument, that specifies path to json file that serves as permanent storage of tasks between sessions.  
`myapp resources/disk.json` . 

Previously loaded task will be pre-loaded into an app and app will run in interactive mode with a number commands available. After `exit` command issued all in-memory state will be dumped into `disk.json` file.

To run tests use `make all` .  
Do `open coverage.html` to see generated coverage report. Report is checked out into repository. 

To build run `make build`.

Check `makefile` for other commands.

App runs similtuneosly as CLI and Web interfaces.


# Code Structure.

Main object is `Task` which gets aggregated into `TasksHolder` with CRUD operations available.  
CLI is responsible for interactive mode, admin access.  
Few packages in a project - internals most of the logic around tasks plus cli, frontend and rest interafaces.
Using `html/templates` to serve front-end on base of standart `net/http` server. Code in view.
Using `embed` package to integrate asset files in a binary, which creates internal read-only file system.
Middleware process context for user access and adds logger wrapper to all handlers.  
Worker pool is implemented but not used.
Task service is transparent.  


# Test
Run `make all` to test with coverage and then `open coverage.html` to see visually appealing report.
I had a git hook in pre-commit for `.git` that runs command each commit, which will contain latest test coverage report.
Aiming for good coverage, happy path and few error handlings.

# Run 
`make build && ./myapp internal/resources/disk.json` . 
To run the app, which will start both CLI and http server on `localhost:8080/tasks`.

# API
Api request can be done like this using http pie for example.   
Create task   
`http POST localhost:8080/api/tasks/4 Authorization:208c0b87-b79e-41fb-a1b3-cd797ef584df Done:=false Msg="Updated Message" Category:=1 PlannedAt="2026-01-02T15:04:05Z"`
Read all tasks  
`http GET localhost:8080/api/tasks`
Get specific task, ids are ints.  
`http GET localhost:8080/api/tasks/{id}`

