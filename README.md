# golang_todo

Microbrewery task managment system.

App should be run with one argument, that specifies path to json file that serves as permanent storage of tasks between sessions.  
`myapp resources/disk.json` . 

Previously loaded task will be pre-loaded into an app and app will run in interactive mode with a number commands available. After `exit` command issued all in-memory state will be dumped into `disk.json` file.

To run tests use `make all` . And then `open coverage.html` to see generated coverage report.

To build run `make build`.

Check `makefile` for other commands.

#Code Structure.

Main object is `Task` which gets aggregated into `TasksHolder` with CRUD operations available.
CLI is responsible for interactive mode.
Package structure is simply flat.  
