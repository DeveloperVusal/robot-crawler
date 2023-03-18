package database

import (
	"context"
	"net/http"
	"time"

	"robot/config"

	"github.com/stevenferrer/solr-go"
)

type Solr struct{}

func (s *Solr) Init() (*solr.JSONClient, string) {
	loadCfg := &config.Database{}
	cfg := loadCfg.Load()

	requestSender := solr.NewDefaultRequestSender().
		WithHTTPClient(&http.Client{Timeout: 10800 * time.Second}).
		WithBasicAuth("solr", "SolrRocks")

	// Create a Solr client
	clientSolr := solr.NewJSONClient(cfg["solr"]["scheme"] + "://" + cfg["solr"]["host"] + ":" + cfg["solr"]["port"]).WithRequestSender(requestSender)

	// check core status
	ctx2 := context.Background()
	_, err2 := clientSolr.CoreStatus(ctx2, solr.NewCoreParams(cfg["solr"]["core"]))

	if err2 != nil {
		panic(err2)
	}

	return clientSolr, cfg["solr"]["core"]
}
