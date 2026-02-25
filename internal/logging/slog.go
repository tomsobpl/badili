package logging

import (
	"log/slog"
	"os"

	"github.com/go-slog/otelslog"
)

func SetupLogger() {
	base := slog.NewJSONHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(otelslog.NewHandler(base)))
}
