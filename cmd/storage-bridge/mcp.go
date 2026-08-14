package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP JSON-RPC Server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMCPServer()
	},
}

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CallToolRequest struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments"`
}

func runMCPServer() error {
	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			sendError(nil, -32700, "Parse error")
			continue
		}

		if req.Method == "tools/call" {
			var toolReq CallToolRequest
			if err := json.Unmarshal(req.Params, &toolReq); err != nil {
				sendError(req.ID, -32602, "Invalid params")
				continue
			}

			handleToolCall(&req, &toolReq)
		} else {
			sendError(req.ID, -32601, "Method not found")
		}
	}
	return scanner.Err()
}

func handleToolCall(req *JSONRPCRequest, toolReq *CallToolRequest) {
	ctx := context.Background()
	var result map[string]interface{}

	switch toolReq.Name {
	case "read_file":
		target := toolReq.Arguments["target"]
		provider, path, err := resolveProvider(target)
		if err != nil {
			sendError(req.ID, -32000, err.Error())
			return
		}
		
		rc, err := provider.Get(ctx, path, 0, -1)
		if err != nil {
			sendError(req.ID, -32000, err.Error())
			return
		}
		defer rc.Close()
		
		data, err := io.ReadAll(rc)
		if err != nil {
			sendError(req.ID, -32000, err.Error())
			return
		}
		
		result = map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": string(data),
				},
			},
		}

	case "write_file":
		target := toolReq.Arguments["target"]
		content := toolReq.Arguments["content"]
		provider, path, err := resolveProvider(target)
		if err != nil {
			sendError(req.ID, -32000, err.Error())
			return
		}

		reader := bytes.NewReader([]byte(content))
		err = provider.Put(ctx, path, reader, int64(len(content)), time.Time{})
		if err != nil {
			sendError(req.ID, -32000, err.Error())
			return
		}

		result = map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "Success",
				},
			},
		}

	case "list_directory":
		target := toolReq.Arguments["target"]
		provider, path, err := resolveProvider(target)
		if err != nil {
			sendError(req.ID, -32000, err.Error())
			return
		}
		
		iter, err := provider.List(ctx, path)
		if err != nil {
			sendError(req.ID, -32000, err.Error())
			return
		}

		var entries []string
		for {
			entry, err := iter.Next(ctx)
			if err == io.EOF {
				break
			}
			if err != nil {
				sendError(req.ID, -32000, err.Error())
				return
			}
			
			if entry.IsDir {
				entries = append(entries, "DIR  "+entry.Path)
			} else {
				entries = append(entries, fmt.Sprintf("%10d %s", entry.Size, entry.Path))
			}
		}

		result = map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": strings.Join(entries, "\n"),
				},
			},
		}

	default:
		sendError(req.ID, -32601, "Tool not found")
		return
	}

	sendResult(req.ID, result)
}

func sendResult(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))
}

func sendError(id interface{}, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))
}
