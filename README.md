# Task Management System

A microservice-based task management system built in Go.

## Overview

This application provides a comprehensive task management solution with multiple interfaces:

- Command Line Interface (CLI)
- Web Interface (localhost:8080)
- REST API (localhost:8080/api)

## Installation

Build the application using:

make build


## Usage

Run the application with:

./myapp internal/resources/tasks.json


The application requires a JSON file path as an argument, which serves as persistent storage between sessions. Tasks are loaded from this file at startup, and all changes are saved to disk when exiting.

## Features

- Multi-interface support (CLI, Web, API)
- Persistent storage
- Interactive command mode
- Concurrent user access
- Task CRUD operations
- Search functionality

## Architecture

### Core Components

- **Task**: Primary data structure
- **TasksHolder**: Aggregator with CRUD operations
- **Interfaces**: CLI, Web Frontend, and REST API
- **Storage**: JSON-based persistence
- **Security**: Middleware for user access and logging
- **Concurrency**: Lock-based task management
- **Worker Pool**: Infrastructure for future analytics

### Technical Details

- Built using standard Go packages (`html/templates`, `net/http`)
- Embedded assets using Go's `embed` package
- Middleware for context processing and logging
- Concurrent access management
- RESTful API implementation

## Testing

Execute the test suite:

make all


View coverage report:

open coverage.html


Automated test coverage reports are generated on each commit via git hooks.

## API Documentation

### Endpoints

#### Create Task

POST localhost:8080/api/tasks/{id}
Authorization: 208c0b87-b79e-41fb-a1b3-cd797ef584df
Content-Type: application/json

{
    "Done": false,
    "Msg": "Task Message",
    "Category": 1,
    "PlannedAt": "2026-01-02T15:04:05Z"
}


#### Read All Tasks

GET localhost:8080/api/tasks


#### Read Specific Task

GET localhost:8080/api/tasks/{id}


## Cloud Infrastructure

The application is designed to work with Google Cloud Storage, automatically persisting in-memory data to Google Cloud Storage buckets when running in Cloud Run containers.

## Development

For additional development commands, refer to the `makefile`.
