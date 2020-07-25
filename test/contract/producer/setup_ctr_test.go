package producer

import (
	"context"
	"github.com/StephanHCB/go-mailer-service/internal/entity"
	"github.com/StephanHCB/go-mailer-service/internal/repository/configuration"
	"github.com/StephanHCB/go-mailer-service/web"
	"github.com/stretchr/testify/mock"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	ts *httptest.Server
)

const tstValidConfigurationPath =  "../../resources/validconfig"

func TestMain(m *testing.M) {
	tstSetup()
	code := m.Run()
	tstShutdown()
	os.Exit(code)
}

func tstSetup() {
	tstSetupConfig()
	tstSetupHttpTestServer()
}

func tstSetupConfig() {
	configuration.SetupForIntegrationTest(func(err error) {}, func(message string) {}, tstValidConfigurationPath, tstValidConfigurationPath)
}

func tstSetupHttpTestServer() {
	server := web.Create()
	web.AddRoutes(server, &MockEmailService{})
	ts = httptest.NewServer(server)
}

func tstShutdown() {
	ts.Close()
}

type MockEmailService struct {
	mock.Mock
}

func (s *MockEmailService) NewInstance(ctx context.Context) *entity.Email {
	return &entity.Email{}
}

func (s *MockEmailService) SendEmail(ctx context.Context, email *entity.Email) error {
	// TODO use mock to verify data for contract tests
	return nil
}
