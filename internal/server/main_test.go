package server

import (
	"context"
	"os"
	"testing"

	"github.com/red-rocket-software/reminder-go/internal/storage"
	"github.com/red-rocket-software/reminder-go/pkg/logging"
)

func newTestServer(store storage.ReminderRepo) *Server {
	logger := logging.GetLogger()

	server := New(context.Background(), logger, store)

	return server
}

func TestMain(m *testing.M) {

	os.Exit(m.Run())
}