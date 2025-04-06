package services

import (
	"context"
	"google.golang.org/grpc/metadata"
	"log/slog"
	"strings"
)

func PrintMetadata(context context.Context) {
	if slog.Default().Enabled(context, slog.LevelDebug) {
		md, ok := metadata.FromIncomingContext(context)
		if ok {
			var args = make([]any, 0)
			for key, values := range md {
				args = append(args, slog.String(key, strings.Join(values, ";")))
			}
			slog.Debug("Metadata", args...)
		}
	}
}
