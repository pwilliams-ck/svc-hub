package main

import (
	"context"

	"github.com/cloudkey-io/service-hub/logger-svc/data"
	"github.com/cloudkey-io/service-hub/logger-svc/logs"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()

	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		res := &logs.LogResponse{Result: "failed"}
		return res, err
	}

	res := &logs.LogResponse{Result: "logged!"}
	return res, nil
}
