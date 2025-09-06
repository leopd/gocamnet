# GoCamNet

A Go executable that demonstrates CLI argument parsing and timing functionality.

## Features

- Counts from 1 to N seconds (where N is a CLI parameter)
- Displays count every second
- Command-line argument parsing with both long and short forms
- Input validation

## Usage

```bash
# Build the executable
go build -o gocamnet

# Run with default 10 seconds
./gocamnet

# Run with custom number of seconds
./gocamnet -seconds 5
./gocamnet -s 15

# Show help
./gocamnet -h
```

## Development

```bash
# Run directly with go
go run main.go -seconds 3

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o gocamnet-linux
GOOS=windows GOARCH=amd64 go build -o gocamnet.exe
```

## Project Structure

```
gocamnet/
├── main.go          # Main application entry point
├── go.mod           # Go module definition
├── .gitignore       # Git ignore patterns
└── README.md        # This file
```

