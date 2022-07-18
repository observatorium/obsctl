package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/bwplotka/mdox/pkg/clilog"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/guptarohit/asciigraph"
	"github.com/observatorium/api/client/models"
	"github.com/observatorium/obsctl/pkg/version"
	"github.com/prometheus/common/model"
	"github.com/spf13/cobra"
	"github.com/wcharczuk/go-chart/v2"
)

const (
	logFormatLogfmt = "logfmt"
	logFormatJson   = "json"
	logFormatCLILog = "clilog"
)

var logLevel, logFormat string
var logger log.Logger

func setupLogger(*cobra.Command, []string) {
	var lvl level.Option
	switch logLevel {
	case "error":
		lvl = level.AllowError()
	case "warn":
		lvl = level.AllowWarn()
	case "info":
		lvl = level.AllowInfo()
	case "debug":
		lvl = level.AllowDebug()
	default:
		panic("unexpected log level")
	}
	switch logFormat {
	case logFormatJson:
		logger = level.NewFilter(log.NewJSONLogger(log.NewSyncWriter(os.Stderr)), lvl)
	case logFormatLogfmt:
		logger = level.NewFilter(log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr)), lvl)
	case logFormatCLILog:
		fallthrough
	default:
		logger = level.NewFilter(clilog.New(log.NewSyncWriter(os.Stderr)), lvl)
	}
}

func NewObsctlCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "obsctl",
		Short:            "CLI to interact with Observatorium",
		Long:             `CLI to interact with Observatorium`,
		Version:          version.Version,
		PersistentPreRun: setupLogger,
	}

	cmd.AddCommand(NewMetricsCmd(ctx))
	cmd.AddCommand(NewContextCommand(ctx))
	cmd.AddCommand(NewLoginCmd(ctx))
	cmd.AddCommand(NewLogoutCmd(ctx))
	cmd.AddCommand(NewTracesCmd(ctx))
	cmd.AddCommand(NewLogsCmd(ctx))

	cmd.PersistentFlags().StringVar(&logLevel, "log.level", "info", "Log filtering level.")
	cmd.PersistentFlags().StringVar(&logFormat, "log.format", logFormatCLILog, "Log format to use.")

	return cmd
}

// prettyPrintJSON prints indented JSON to stdout.
func prettyPrintJSON(b []byte) (string, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "\t")
	if err != nil {
		level.Debug(logger).Log("msg", "failed indent", "json", string(b))
		return "", fmt.Errorf("indent JSON %w", err)
	}

	return out.String(), nil
}

func handleResponse(body []byte, contentType string, statusCode int, cmd *cobra.Command) error {
	if statusCode/100 == 2 {
		json, err := prettyPrintJSON(body)
		if err != nil {
			return fmt.Errorf("request failed with status code %d pretty printing: %v", statusCode, err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), json)
		return nil
	}

	if len(body) != 0 {
		// Pretty print only if we know the error response is JSON.
		// In future we might want to handle other types as well.
		switch contentType {
		case "application/json":
			jsonErr, err := prettyPrintJSON(body)
			if err != nil {
				return fmt.Errorf("request failed with status code %d pretty printing: %v", statusCode, err)
			}

			return fmt.Errorf(jsonErr)
		default:
			return fmt.Errorf("request failed with status code %d, error: %s", statusCode, string(body))
		}
	}

	return fmt.Errorf("request failed with status code %d", statusCode)
}

func handleGraph(body []byte, graph, query string, cmd *cobra.Command) error {
	// TODO(saswatamcode): Update spec so that we can use client/models directly.
	var m struct {
		Data struct {
			ResultType string          `json:"resultType"`
			Result     json.RawMessage `json:"result"`
		} `json:"data"`

		Error     string `json:"error,omitempty"`
		ErrorType string `json:"errorType,omitempty"`
		// Extra field supported by Thanos Querier.
		Warnings []string `json:"warnings"`
	}

	if err := json.Unmarshal(body, &m); err != nil {
		return fmt.Errorf("unmarshal query range response %w", err)
	}

	var matrixResult model.Matrix

	// Decode the Result depending on the ResultType
	switch m.Data.ResultType {
	case string(models.RangeQueryResponseResultTypeMatrix):
		if err := json.Unmarshal(m.Data.Result, &matrixResult); err != nil {
			return fmt.Errorf("decode result into ValueTypeMatrix %w", err)
		}
	default:
		if m.Warnings != nil {
			return fmt.Errorf("error: %s, type: %s, warning: %s", m.Error, m.ErrorType, strings.Join(m.Warnings, ", "))
		}
		if m.Error != "" {
			return fmt.Errorf("error: %s, type: %s", m.Error, m.ErrorType)
		}

		return fmt.Errorf("received status code: 200, unknown response type: '%q'", m.Data.ResultType)
	}

	// Output graph based on type specified.
	switch graph {
	case "ascii":
		var data [][]float64

		for _, ss := range matrixResult {
			stream := []float64{}
			for _, sample := range ss.Values {
				stream = append(stream, float64(sample.Value))
			}
			data = append(data, stream)
		}

		// TODO(saswatamcode): Output data in some format and use standard graphing tools.
		fmt.Fprintln(cmd.OutOrStdout(), asciigraph.PlotMany(data, asciigraph.Width(80)))
		return nil
	case "png":
		var data []chart.Series

		for _, ss := range matrixResult {
			xstream := []time.Time{}
			ystream := []float64{}
			for _, sample := range ss.Values {
				ystream = append(ystream, float64(sample.Value))
				xstream = append(xstream, sample.Timestamp.Time())
			}
			data = append(data, chart.TimeSeries{
				Name:    ss.Metric.String(),
				XValues: xstream,
				YValues: ystream,
				YAxis:   chart.YAxisPrimary,
			})
		}

		graph := chart.Chart{
			XAxis: chart.XAxis{
				Name: "Time",
			},
			YAxis: chart.YAxis{
				Name: "Value",
			},
			Series: data,
		}

		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get working dir: %w", err)
		}

		f, err := os.Create(path.Join(wd, "graph"+time.Now().String()+".png"))
		if err != nil {
			return fmt.Errorf("could not create graph png file: %w", err)
		}
		defer f.Close()

		if err := graph.Render(chart.PNG, f); err != nil {
			return fmt.Errorf("could not render graph: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("unsupported graph type: %s", graph)
	}
}
