package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Count   int         `json:"count,omitempty"`
}

var dbMutexes = struct {
	sync.Mutex
	m map[string]*sync.Mutex
}{m: make(map[string]*sync.Mutex)}

func getDBMutex(db string) *sync.Mutex {
	dbMutexes.Lock()
	defer dbMutexes.Unlock()

	if _, exists := dbMutexes.m[db]; !exists {
		dbMutexes.m[db] = &sync.Mutex{}
	}
	return dbMutexes.m[db]
}

func writeJSON(conn net.Conn, resp Response) {
	b, _ := json.Marshal(resp)
	conn.Write(append(b, '\n'))
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 4)
		if len(parts) != 4 {
			writeJSON(conn, Response{
				Status:  "error",
				Message: "invalid command format",
			})
			continue
		}

		database, collection, action, jsonArg := parts[0], parts[1], parts[2], parts[3]

		if len(jsonArg) > 1 {
			if (jsonArg[0] == '"' && jsonArg[len(jsonArg)-1] == '"') ||
				(jsonArg[0] == '\'' && jsonArg[len(jsonArg)-1] == '\'') {
				jsonArg = jsonArg[1 : len(jsonArg)-1]
			}
		}

		dbMutex := getDBMutex(database)
		fmt.Printf("Client %v: waiting for lock on database %s\n", conn.RemoteAddr(), database)
		dbMutex.Lock()
		fmt.Printf("Client %v: acquired lock on database %s\n", conn.RemoteAddr(), database)

		cmd := exec.Command("../database/main", database, collection, action, jsonArg)
		output, err := cmd.CombinedOutput() // []byte
		dbMutex.Unlock()
		fmt.Printf("Client %v: released lock on database %s\n", conn.RemoteAddr(), database)

		if err != nil {
			writeJSON(conn, Response{
				Status:  "error",
				Message: err.Error(),
			})
			continue
		}

		outputStr := strings.TrimSpace(string(output))

		switch action {
		case "find":
			var data []interface{}
			_ = json.Unmarshal([]byte(outputStr), &data)
			writeJSON(conn, Response{
				Status:  "success",
				Message: fmt.Sprintf("Fetched %d docs from %s", len(data), collection),
				Data:    data,
				Count:   len(data),
			})
		case "insert":
			writeJSON(conn, Response{
				Status:  "success",
				Message: fmt.Sprintf("Document inserted into %s", collection),
			})
		case "delete":
			var data []interface{}
			_ = json.Unmarshal([]byte(outputStr), &data)
			writeJSON(conn, Response{
				Status:  "success",
				Message: fmt.Sprintf("Deleted %d documents from %s", len(data), collection),
				Count:   len(data),
			})
		default:
			writeJSON(conn, Response{
				Status:  "error",
				Message: "unknown action",
			})
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	fmt.Println("Server started on :8080")
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}
