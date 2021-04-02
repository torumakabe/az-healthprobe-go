package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/logutils"
	"github.com/jszwec/csvutil"
	"github.com/kelseyhightower/envconfig"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/robfig/cron/v3"
)

type Config struct {
	Location       string `required:"true" envconfig:"probe_location"`
	IKey           string `required:"true" envconfig:"probe_instrumentation_key"`
	LogLevel       string `default:"INFO" envconfig:"probe_log_level"`
	TargetListFile string `default:"./conf/sample_target.csv" envconfig:"probe_target_list_file"`
}

type ProbeTarget struct {
	Name      string `csv:"target_name"`
	Url       string `csv:"target_url"`
	Frequency int    `csv:"frequency"`
}

type ProbeResult struct {
	HttpStatus string
	StartTime  time.Time
	EndTime    time.Time
}

var (
	conf            Config
	telemetryClient appinsights.TelemetryClient
)

func probe(url string) (*ProbeResult, error) {
	result := &ProbeResult{}
	result.StartTime = time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to get: %s error: %s", url, err)
	}
	result.EndTime = time.Now()
	resp.Body.Close()
	result.HttpStatus = resp.Status

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("invalid status: %s status: %s", url, resp.Status)
	}

	return result, nil
}

func sendAvailablity(testName string, duration time.Duration, success bool, msg string, location string) error {
	availability := appinsights.NewAvailabilityTelemetry(testName, duration, success)
	availability.RunLocation = location
	availability.Message = msg
	telemetryClient.Track(availability)
	telemetryClient.Channel().Flush()

	// placeholder (telemetryClient does not return any value so far)
	return nil
}

func probeInvoker(name, url string) {
	duration, err := time.ParseDuration("0")
	if err != nil {
		log.Fatalf("[ERROR] %s", err.Error())
		return
	}

	var (
		success bool
		msg     string
	)
	result, err := probe(url)
	if err != nil {
		msg = err.Error()
	} else {
		duration = result.EndTime.Sub(result.StartTime)
		success = true
		msg = result.HttpStatus
	}

	testName := fmt.Sprintf("Health Probe [Target: %s]", name)
	if err := sendAvailablity(testName, duration, success, msg, conf.Location); err != nil {
		log.Printf("[ERROR] %s", err.Error())
	}

	log.Printf("[DEBUG] Health Probe [Target: %s]: %s", name, msg)
}

func Run() int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	err := envconfig.Process("", &conf)
	if err != nil {
		log.Fatal(err.Error())
	}

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(conf.LogLevel),
		Writer:   os.Stdout,
	}
	log.SetOutput(filter)

	telemetryConfig := appinsights.NewTelemetryConfiguration(conf.IKey)
	telemetryClient = appinsights.NewTelemetryClientFromConfig(telemetryConfig)

	appinsights.NewDiagnosticsMessageListener(func(msg string) error {
		log.Printf("[DEBUG] %s", msg)
		return nil
	})

	var targets []ProbeTarget
	b, _ := ioutil.ReadFile(conf.TargetListFile)
	if err := csvutil.Unmarshal(b, &targets); err != nil {
		log.Fatal(err.Error())
	}

	c := cron.New()

	for _, target := range targets {
		freq := fmt.Sprintf("@every %ds", target.Frequency)
		log.Printf("[INFO] registering probe... %s, %s, %s", target.Name, target.Url, freq)
		go func(n, u, f string) {
			c.AddFunc(f, func() {
				probeInvoker(n, u)
			})
		}(target.Name, target.Url, freq)
	}

	c.Start()
	log.Print("[INFO] Probe(s) started")

	<-ctx.Done()
	log.Printf("[WARN] got signal: %s", ctx.Err())
	c.Stop()

	select {
	case <-telemetryClient.Channel().Close(5 * time.Second):
		log.Print("[INFO] closed telemetry channel successfully")
	case <-time.After(8 * time.Second):
		log.Print("[WARN] some telemetry may not have been sent")
	}

	log.Print("[INFO] Probe(s) was terminated")
	return 0

}

func main() { os.Exit(Run()) }
