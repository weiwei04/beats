package pandora

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/common/op"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/outputs"
	psdk "qiniu.com/pandora/pipeline"
)

func init() {
	outputs.RegisterOutputPlugin("pandora", New)
}

type pandora struct {
	client   psdk.PipelineAPI
	repo     string
	hostname string
	batch    int
	pointBuf []psdk.Point
	retries  int
}

//func New(beatName string, cfg *common.Config, _ int) (outputs.Outputer, error) {
func New(beatInfo common.BeatInfo, cfg *common.Config, _ int) (outputs.Outputer, error) {
	config := defaultConfig
	err := cfg.Unpack(&config)
	if err != nil {
		logp.Err("unpack config failed, err[%s]", err)
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		logp.Err("failed to get hostname, err[%s]", err)
		return nil, err
	}

	pconfig := psdk.NewConfig().WithEndpoint(config.Endpoint).WithAccessKeySecretKey(config.AK, config.SK)
	pclient, err := psdk.New(pconfig)
	if err != nil {
		logp.Err("create pandora client failed, err[%s]", err)
		return nil, err
	}

	repo := fmt.Sprintf("dora_%s", config.Region)

	return &pandora{
		client:   pclient,
		repo:     repo,
		hostname: hostname,
		batch:    config.Batch,
		pointBuf: make([]psdk.Point, 0, config.Batch),
		retries:  config.MaxRetries,
	}, nil
}

func (p *pandora) PublishEvent(
	sig op.Signaler,
	opts outputs.Options,
	data outputs.Data,
) error {
	if len(p.pointBuf) >= p.batch {
		p.sendPoints()
	}
	p.pointBuf = append(p.pointBuf, convertToPoint(p.hostname, data.Event))
	op.Sig(sig, nil)
	return nil
}

func (p *pandora) sendPoints() error {
	points := &psdk.PostDataInput{
		RepoName: p.repo,
		Points:   psdk.Points(p.pointBuf),
	}
	for i := 0; i < p.retries; i++ {
		if err := p.client.PostData(points); err != nil {
			logp.Err("post data failed at try %d, err[%s]", i, err)
		} else {
			logp.Info("published %d points", len(p.pointBuf))
			break
		}
	}
	p.pointBuf = p.pointBuf[:0]
	return nil
}

func (p *pandora) Close() error {
	if len(p.pointBuf) > 0 {
		p.sendPoints()
	}
	return nil
}

func escapeString(str string) string {
	newStr := strings.Replace(str, `\`, `\\`, -1)
	newStr = strings.Replace(newStr, "\t", `\\t`, -1)
	newStr = strings.Replace(newStr, "\r", `\\r`, -1)
	newStr = strings.Replace(newStr, "\n", `\\n`, -1)
	return newStr
}

func mapStrToSlice(hostname string, event common.MapStr) []psdk.PointField {
	fields := []psdk.PointField{}

	fields = append(fields, psdk.PointField{Key: "hostname", Value: hostname})

	message := event["message"].(string)
	fields = append(fields, psdk.PointField{Key: "message", Value: escapeString(message)})

	ts := event["@timestamp"].(common.Time)
	fields = append(fields, psdk.PointField{Key: "timestamp", Value: time.Time(ts).Format(time.RFC3339)})

	logType := event["type"].(string)
	if logType == "stdout" ||
		logType == "stderr" ||
		logType == "sandbox" {
		path := event["source"].(string)
		parts := strings.Split(path, "/")
		var i int
		var p string
		for i, p = range parts {
			if p == "executors" {
				executorId := parts[i+1]
				executorIdParts := strings.Split(parts[i+1], ".")
				fields = append(fields, psdk.PointField{Key: "instance_id", Value: executorId})
				fields = append(fields, psdk.PointField{Key: "app", Value: executorIdParts[0]})
				fields = append(fields, psdk.PointField{Key: "launch_id", Value: executorIdParts[1]})
			}
		}

		if logType == "sandbox" {
			logType = parts[i-1]
		}
		fields = append(fields, psdk.PointField{Key: "source", Value: logType})
	}

	return fields
}

func convertToPoint(hostName string, event common.MapStr) psdk.Point {
	fields := mapStrToSlice(hostName, event)
	return psdk.Point{Fields: fields}
}
