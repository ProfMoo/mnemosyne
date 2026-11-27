package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/profmoo/mnemosyne/gen/api"
	"github.com/profmoo/mnemosyne/internal/musicbrainz"
	"github.com/profmoo/mnemosyne/internal/server"
)

type LoggingSuite struct {
	suite.Suite

	logs bytes.Buffer
}

type loggingRecord struct {
	Message      string          `json:"msg"`
	Level        string          `json:"level"`
	Method       string          `json:"method"`
	Path         string          `json:"path"`
	Operation    string          `json:"operation"`
	Payload      string          `json:"payload"`
	PayloadError string          `json:"payload_error"`
	Error        string          `json:"error"`
	Duration     json.RawMessage `json:"duration"`
}

type loggingExchange struct {
	request  *loggingRecord
	response *loggingRecord
}

//nolint:paralleltest // These tests replace the process-wide default logger.
func TestLoggingSuite(t *testing.T) {
	suite.Run(t, new(LoggingSuite))
}

func (s *LoggingSuite) SetupTest() {
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&s.logs, nil)))
	s.T().Cleanup(func() { slog.SetDefault(previous) })
	s.logs.Reset()
}

func (s *LoggingSuite) TestRequestResponseLogs() {
	input, err := os.ReadFile("../../testdata/single/01-minimal.json")
	s.Require().NoError(err)

	body := new(api.SubmitListensJSONRequestBody)
	s.Require().NoError(json.Unmarshal(input, body))
	request := &api.SubmitListensRequestObject{Body: body}
	response := &api.SubmitListens201JSONResponse{Status: "ok"}

	logs := s.invoke(request, response, nil)
	requestLog, responseLog := logs.request, logs.response
	s.Equal("request", requestLog.Message)
	s.Equal("response", responseLog.Message)
	s.Equal("INFO", responseLog.Level)
	s.NotEmpty(responseLog.Duration)
	s.Empty(responseLog.Error)

	for _, record := range []*loggingRecord{requestLog, responseLog} {
		s.Equal(http.MethodPost, record.Method)
		s.Equal("/1/submit-listens", record.Path)
		s.Equal("submitListens", record.Operation)
	}

	expected, err := json.Marshal(request)
	s.Require().NoError(err)
	s.JSONEq(string(expected), requestLog.Payload)
	s.JSONEq(`{"status":"ok"}`, responseLog.Payload)
	s.NotContains(s.logs.String(), "secret-token")
	s.NotContains(s.logs.String(), "secret-query")
}

func (s *LoggingSuite) TestHandlerError() {
	upstreamErr := &musicbrainz.HTTPError{StatusCode: http.StatusServiceUnavailable}
	logs := s.invoke(&api.SubmitListensRequestObject{}, nil, upstreamErr)
	responseLog := logs.response

	s.Equal("response", responseLog.Message)
	s.Equal("ERROR", responseLog.Level)
	s.Equal(upstreamErr.Error(), responseLog.Error)
	s.Equal("null", responseLog.Payload)
}

func (s *LoggingSuite) TestUnserializablePayloadDoesNotChangeResult() {
	payload := make(chan struct{})
	logs := s.invoke(payload, payload, nil)

	for _, record := range []*loggingRecord{logs.request, logs.response} {
		s.NotEmpty(record.PayloadError)
		s.Empty(record.Payload)
	}
}

func (s *LoggingSuite) invoke(request, response any, handlerErr error) loggingExchange {
	s.T().Helper()

	httpRequest := httptest.NewRequestWithContext(s.T().Context(), http.MethodPost,
		"/1/submit-listens?token=secret-query", http.NoBody)
	httpRequest.Header.Set("Authorization", "Token secret-token")
	recorder := httptest.NewRecorder()
	handler := server.NewLoggingMiddleware()(
		func(_ context.Context, _ http.ResponseWriter, _ *http.Request, input any) (any, error) {
			s.Equal(request, input)
			return response, handlerErr
		},
		"submitListens",
	)

	actual, err := handler(httpRequest.Context(), recorder, httpRequest, request)
	s.Equal(response, actual)
	if handlerErr != nil {
		s.Require().ErrorIs(err, handlerErr)
	} else {
		s.Require().NoError(err)
	}

	var requestLog, responseLog, extra loggingRecord
	decoder := json.NewDecoder(bytes.NewReader(s.logs.Bytes()))
	s.Require().NoError(decoder.Decode(&requestLog))
	s.Require().NoError(decoder.Decode(&responseLog))
	s.Require().ErrorIs(decoder.Decode(&extra), io.EOF)

	return loggingExchange{request: &requestLog, response: &responseLog}
}
