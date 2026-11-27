package api_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/profmoo/mnemosyne/gen/api"
	"github.com/profmoo/mnemosyne/gen/client"
)

type SubmitListensSuite struct {
	suite.Suite
}

func TestSubmitListensSuite(t *testing.T) {
	suite.Run(t, new(SubmitListensSuite))
}

func (s *SubmitListensSuite) TestRequestBodyRoundTrip() {
	fixtures, err := filepath.Glob("../../testdata/*/*.json")
	s.Require().NoError(err)
	s.Require().NotEmpty(fixtures)

	for _, implementation := range []struct {
		name    string
		newBody func() any
	}{
		{name: "server", newBody: func() any { return new(api.SubmitListensJSONRequestBody) }},
		{name: "client", newBody: func() any { return new(client.SubmitListensJSONRequestBody) }},
	} {
		s.Run(implementation.name, func() {
			for _, fixture := range fixtures {
				s.Run(filepath.Join(filepath.Base(filepath.Dir(fixture)), filepath.Base(fixture)), func() {
					input, err := os.ReadFile(fixture)
					s.Require().NoError(err)

					body := implementation.newBody()
					s.Require().NoError(json.Unmarshal(input, body))

					output, err := json.Marshal(body)
					s.Require().NoError(err)
					s.JSONEq(string(input), string(output))
				})
			}
		})
	}
}
