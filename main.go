package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"gitlab.com/pickledrick/g8s-scaler/g8s"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const app = "g8s-autoscaler"

var c Config

type Config struct {
	Token       string `required:"true"`
	Cluster     string `required:"true"`
	FetchURL    string `required:"true"`
	FetchPath   string `default:"/metrics"`
	FetchMetric string `default:"required_nodes"`
	G8sAPI      string `required:"true"`
	Logger      log.Logger
}

func parseMetric(data string, metric string) string {
	for _, l := range strings.Split(data, "\n") {
		if strings.HasPrefix(l, metric) {
			m := strings.Split(l, " ")
			return m[len(m)-1]
		}
	}
	return "none"
}

func init() {

	c.Logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	c.Logger = log.With(c.Logger, "ts", log.DefaultTimestampUTC, "app", app)

	c.Token = os.Getenv("TOKEN")
	c.Cluster = os.Getenv("CLUSTER")
	c.FetchURL = os.Getenv("FETCH_URL")
	c.FetchPath = os.Getenv("FETCH_PATH")
	c.FetchMetric = os.Getenv("FETCH_METRIC")
	c.G8sAPI = os.Getenv("G8S_API")

}

func main() {

	c.Logger.Log("msg", "starting")

	desired, err := fetchDesired(&c)
	if err != nil {
		level.Error(c.Logger).Log("msg", "failed to fetch desired nodes", "error", err)
		os.Exit(1)
	}
	if desired < 1 {
		c.Logger.Log("msg", "unable to scale below 1")
		os.Exit(1)
	}

	updateWorkers(&c, desired)

}

func fetchDesired(*Config) (int, error) {
	res, err := http.Get(c.FetchURL + c.FetchPath)
	if err != nil {
		level.Error(c.Logger).Log("msg", "failed to fetch desired")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(res.Body)
		if err != nil {
			return 0, errors.Wrapf(err, "request fetch desired nodes failed with [%d], but the response body could not be read", res.StatusCode)
		}
		return 0, errors.Errorf("request fetch desired nodes failed, code [%d], url [%s], body: %s", res.StatusCode, c.FetchURL+c.FetchPath, buf.String())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	required := parseMetric(string(body), c.FetchMetric)
	numreq, err := strconv.Atoi(required)
	if err != nil {
		return 0, errors.Wrapf(err, "unable to convert desired nodes to integer from metric")
	}
	return numreq, nil
}

func fetchCluster(c *Config) (*g8s.Cluster, error) {
	var cluster g8s.Cluster
	url := c.G8sAPI + c.Cluster + "/"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "giantswarm "+c.Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &cluster, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(res.Body)
		if err != nil {
			return &cluster, errors.Wrapf(err, "request fetch cluster [%s] failed with [%d], but the response body could not be read", c.Cluster, res.StatusCode)
		}
		return &cluster, errors.Errorf("request fetch cluster [%s] failed, code [%d], url [%s], body: %s", res.StatusCode, url, buf.String())
	}

	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)

	err = dec.Decode(&cluster)
	if err != nil {
		level.Error(c.Logger).Log("msg", "failed to decode cluster")
	}

	return &cluster, nil
}

func updateWorkers(c *Config, desired int) {
	cluster, err := fetchCluster(c)
	if err != nil {
		level.Error(c.Logger).Log("msg", "failed to fetch cluster information", "error", err)
		os.Exit(1)
	}

	current := len(cluster.Workers)

	c.Logger.Log("msg", "node count", "current", current)
	c.Logger.Log("msg", "node count", "desired", desired)

	var new g8s.Cluster

	new.Workers = cluster.Workers[:len(cluster.Workers)-1]

	switch {
	case current > desired:
		c.Logger.Log("msg", "scaling cluster down")
		diff := len(cluster.Workers) - desired
		new.Workers = cluster.Workers[:len(cluster.Workers)-diff]
		patchCluster(c, &new)
	case current < desired:
		c.Logger.Log("msg", "scaling cluster up")
		diff := desired - len(cluster.Workers)
		new.Workers = cluster.Workers
		for i := 1; i <= diff; i++ {
			var worker g8s.Worker
			new.Workers = append(new.Workers, worker)
		}
		patchCluster(c, &new)
	case current == desired:
		c.Logger.Log("msg", "no scaling required")
	}
}

func patchCluster(c *Config, clusterSpec *g8s.Cluster) error {
	data, err := json.Marshal(clusterSpec)
	if err != nil {
		return err
	}
	url := c.G8sAPI + c.Cluster + "/"
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "giantswarm "+c.Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(res.Body)
		if err != nil {
			return errors.Wrapf(err, "request to update cluster [%s] failed with [%d], but the response body could not be read", c.Cluster, res.StatusCode)
		}
		return errors.Errorf("request to update cluster cluster [%s] failed, code [%d], url [%s], body: %s", res.StatusCode, url, buf.String())
	}
	return nil
}
