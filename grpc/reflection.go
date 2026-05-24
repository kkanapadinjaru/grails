package grpc

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// MethodDescription represents the description of a gRPC method including request and response types
type MethodDescription struct {
	Name         string
	RequestType  string
	ResponseType string
	RequestJSON  string // Pre-generated skeleton JSON
}

func runGrpcurl(args ...string) (string, error) {
	// Add timeout context to prevent hanging
	cmd := exec.Command("grpcurl", args...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("grpcurl error: %v, stderr: %s", err, stderr.String())
	}
	
	return stdout.String(), nil
}

// ListServices lists all gRPC services available on the given address
func ListServices(address string) ([]string, error) {
	log.Printf("[ListServices] Running grpcurl to list services on %s", address)
	
	output, err := runGrpcurl("-plaintext", address, "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	
	var services []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && line != "grpc.reflection.v1alpha.ServerReflection" && line != "grpc.reflection.v1.ServerReflection" {
			services = append(services, line)
		}
	}
	
	return services, nil
}

// GetGrpcMethods lists all methods available on a specific gRPC service
func GetGrpcMethods(address, serviceName string) ([]string, error) {
	log.Printf("[GetGrpcMethods] Running grpcurl to list methods for service %s", serviceName)
	
	output, err := runGrpcurl("-plaintext", address, "list", serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to list methods: %w", err)
	}
	
	var methods []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			// Extract just the method name
			parts := strings.Split(line, ".")
			methodName := parts[len(parts)-1]
			methods = append(methods, methodName)
		}
	}
	
	return methods, nil
}

// DescribeMethod describes a specific gRPC method to get its request and response types
func DescribeMethod(address, serviceName, methodName string) (*MethodDescription, error) {
	fullMethodName := fmt.Sprintf("%s.%s", serviceName, methodName)
	log.Printf("[DescribeMethod] Running grpcurl to describe method %s", fullMethodName)
	
	output, err := runGrpcurl("-plaintext", address, "describe", fullMethodName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe method: %w", err)
	}
	
	// Output looks like:
	// chinook.MusicService.GetArtistById is a method:
	// rpc GetArtistById ( .chinook.GetRequest ) returns ( .chinook.Artist );
	
	reqType := ""
	respType := ""
	
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "rpc ") {
			// Extract request type
			startReq := strings.Index(line, "(")
			endReq := strings.Index(line, ")")
			if startReq != -1 && endReq != -1 {
				reqType = strings.TrimSpace(line[startReq+1 : endReq])
				reqType = strings.TrimPrefix(reqType, ".")
			}
			
			// Extract response type
			startResp := strings.LastIndex(line, "(")
			endResp := strings.LastIndex(line, ")")
			if startResp != -1 && endResp != -1 && startResp > endReq {
				respType = strings.TrimSpace(line[startResp+1 : endResp])
				respType = strings.TrimPrefix(respType, ".")
			}
		}
	}
	
	if reqType == "" || respType == "" {
		return nil, fmt.Errorf("failed to parse method description from output: %s", output)
	}
	
	return &MethodDescription{
		Name:         methodName,
		RequestType:  reqType,
		ResponseType: respType,
	}, nil
}

// GenerateJsonSkeleton generates a JSON skeleton from a message description
func GenerateJsonSkeleton(address, messageType string) (string, error) {
	log.Printf("[GenerateJsonSkeleton] Running grpcurl to get msg-template for %s", messageType)
	
	output, err := runGrpcurl("-msg-template", "-plaintext", address, "describe", messageType)
	if err != nil {
		return "", fmt.Errorf("failed to generate JSON skeleton: %w", err)
	}
	
	// Find the "Message template:" section
	templateIdx := strings.Index(output, "Message template:")
	if templateIdx == -1 {
		return "{}", nil
	}
	
	jsonOutput := strings.TrimSpace(output[templateIdx+len("Message template:"):])
	return jsonOutput, nil
}

// SendGrpcRequest sends a request to the given gRPC method and returns the response JSON
func SendGrpcRequest(address, serviceName, methodName, requestJSON, bearerToken string) (string, error) {
	fullMethodName := fmt.Sprintf("%s.%s", serviceName, methodName)
	log.Printf("[SendGrpcRequest] Sending request to %s via grpcurl", fullMethodName)
	
	args := []string{"-plaintext"}
	
	if bearerToken != "" {
		args = append(args, "-H", "Authorization: Bearer "+bearerToken)
	}
	
	args = append(args, "-d", requestJSON)
	args = append(args, address)
	args = append(args, fullMethodName)
	
	// Add timeout context to prevent hanging
	cmd := exec.Command("grpcurl", args...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Let's use a timeout for safety
	errChan := make(chan error, 1)
	go func() {
		errChan <- cmd.Run()
	}()
	
	select {
	case <-time.After(30 * time.Second):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return "", fmt.Errorf("grpcurl request timed out after 30s")
	case err := <-errChan:
		if err != nil {
			// Check if stderr contains a useful message
			stderrStr := stderr.String()
			if strings.Contains(stderrStr, "ERROR:") {
				// Parse grpcurl error output
				// ERROR:
				//   Code: Unknown
				//   Message: Exception was thrown by handler.
				msgIdx := strings.Index(stderrStr, "Message: ")
				if msgIdx != -1 {
					errMsg := strings.TrimSpace(stderrStr[msgIdx+len("Message: "):])
					// Return the clean error message
					return "", fmt.Errorf("rpc error: %s", errMsg)
				}
				return "", fmt.Errorf("grpc error: %s", stderrStr)
			}
			return "", fmt.Errorf("grpcurl execution failed: %v, stderr: %s", err, stderrStr)
		}
		return stdout.String(), nil
	}
}
