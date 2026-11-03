package db

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
)

func TestPostNodeJobToBeaconNodeNormalizesEndpoint(t *testing.T) {
	var gotPath string
	var gotContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer server.Close()

	utils.Config = &types.Config{}
	utils.Config.NodeJobsProcessor.ClEndpoint = server.URL + "///"

	statusCode, status, responseBody, err := postNodeJobToBeaconNode(
		"/eth/v1/beacon/pool/voluntary_exits",
		[]byte(`{"message":{}}`),
	)
	if err != nil {
		t.Fatalf("postNodeJobToBeaconNode returned an error: %v", err)
	}
	if statusCode != http.StatusAccepted {
		t.Fatalf("status code = %d, want %d", statusCode, http.StatusAccepted)
	}
	if status != "202 Accepted" {
		t.Fatalf("status = %q, want %q", status, "202 Accepted")
	}
	if gotPath != "/eth/v1/beacon/pool/voluntary_exits" {
		t.Fatalf("request path = %q, want normalized beacon API path", gotPath)
	}
	if gotContentType != "application/json" {
		t.Fatalf("content type = %q, want application/json", gotContentType)
	}
	if string(responseBody) != `{"data":{}}` {
		t.Fatalf("response body = %q, want beacon response body", responseBody)
	}
}

func TestPostNodeJobToBeaconNodeRequiresEndpoint(t *testing.T) {
	utils.Config = &types.Config{}

	_, _, _, err := postNodeJobToBeaconNode("/eth/v1/beacon/pool/voluntary_exits", nil)
	if err == nil || !strings.Contains(err.Error(), "endpoint is not configured") {
		t.Fatalf("error = %v, want missing endpoint error", err)
	}
}