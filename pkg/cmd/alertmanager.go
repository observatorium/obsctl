package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/observatorium/api/client"
	"github.com/observatorium/api/client/models"
	"github.com/observatorium/obsctl/pkg/fetcher"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/prometheus/alertmanager/pkg/labels"
	"github.com/spf13/cobra"
)

func NewAlertmanagerGetCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get alerts or silences from Alertmanager.",
		Long:  "Get alerts or silences from Alertmanager for a tenant.",
	}

	// Alerts command.
	var (
		activeFilter       bool
		silencedFilter     bool
		inhibitedFilter    bool
		unprocessedFilter  bool
		receiverFilter     string
		labelFilters       []string
		alertsOutputFormat string
	)
	alertsCmd := &cobra.Command{
		Use:          "alerts",
		Short:        "Get alerts from Alertmanager.",
		Long:         "Get alerts from Alertmanager for a tenant.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			params := &client.GetAlertsParams{}
			if cmd.Flags().Changed("active") {
				params.Active = &activeFilter
			}
			if cmd.Flags().Changed("silenced") {
				params.Silenced = &silencedFilter
			}
			if cmd.Flags().Changed("inhibited") {
				params.Inhibited = &inhibitedFilter
			}
			if cmd.Flags().Changed("unprocessed") {
				params.Unprocessed = &unprocessedFilter
			}
			if receiverFilter != "" {
				params.Receiver = &receiverFilter
			}
			if len(labelFilters) > 0 {
				params.Filter = &labelFilters
			}

			resp, err := f.GetAlertsWithResponse(ctx, currentTenant, params)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			if resp.StatusCode()/100 != 2 {
				return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
			}

			if alertsOutputFormat == "table" {
				var alerts models.GettableAlerts
				if err := json.Unmarshal(resp.Body, &alerts); err != nil {
					return fmt.Errorf("parsing alerts: %w", err)
				}
				renderAlertsTable(cmd.OutOrStdout(), alerts)
				return nil
			}

			return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
		},
	}
	alertsCmd.Flags().BoolVar(&activeFilter, "active", false, "Show active alerts")
	alertsCmd.Flags().BoolVar(&silencedFilter, "silenced", false, "Show silenced alerts")
	alertsCmd.Flags().BoolVar(&inhibitedFilter, "inhibited", false, "Show inhibited alerts")
	alertsCmd.Flags().BoolVar(&unprocessedFilter, "unprocessed", false, "Show unprocessed alerts")
	alertsCmd.Flags().StringVar(&receiverFilter, "receiver", "", "Regex filter by receiver name")
	alertsCmd.Flags().StringArrayVarP(&labelFilters, "filter", "f", nil, "Filter alerts by label matchers (e.g., alertname=HighCPU)")
	alertsCmd.Flags().StringVarP(&alertsOutputFormat, "output", "o", "json", "Output format: json or table")

	// Silences command.
	var (
		silenceLabelFilters   []string
		silencesOutputFormat  string
	)
	silencesCmd := &cobra.Command{
		Use:          "silences",
		Short:        "Get silences from Alertmanager.",
		Long:         "Get silences from Alertmanager for a tenant.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			params := &client.GetSilencesParams{}
			if len(silenceLabelFilters) > 0 {
				params.Filter = &silenceLabelFilters
			}

			resp, err := f.GetSilencesWithResponse(ctx, currentTenant, params)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			if resp.StatusCode()/100 != 2 {
				return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
			}

			if silencesOutputFormat == "table" {
				var silences models.GettableSilences
				if err := json.Unmarshal(resp.Body, &silences); err != nil {
					return fmt.Errorf("parsing silences: %w", err)
				}
				renderSilencesTable(cmd.OutOrStdout(), silences)
				return nil
			}

			return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
		},
	}
	silencesCmd.Flags().StringArrayVarP(&silenceLabelFilters, "filter", "f", nil, "Filter silences by label matchers")
	silencesCmd.Flags().StringVarP(&silencesOutputFormat, "output", "o", "json", "Output format: json or table")

	// Individual silence command.
	var silenceOutputFormat string
	silenceCmd := &cobra.Command{
		Use:          "silence [silenceID]",
		Short:        "Get a specific silence from Alertmanager.",
		Long:         "Get a specific silence from Alertmanager by its ID.",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			silenceID, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid silence ID: %w", err)
			}
			resp, err := f.GetSilenceWithResponse(ctx, currentTenant, silenceID)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			if resp.StatusCode()/100 != 2 {
				return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
			}

			if silenceOutputFormat == "table" {
				var silence models.GettableSilence
				if err := json.Unmarshal(resp.Body, &silence); err != nil {
					return fmt.Errorf("parsing silence: %w", err)
				}
				renderSilenceTable(cmd.OutOrStdout(), silence)
				return nil
			}

			return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
		},
	}
	silenceCmd.Flags().StringVarP(&silenceOutputFormat, "output", "o", "json", "Output format: json or table")

	cmd.AddCommand(alertsCmd)
	cmd.AddCommand(silencesCmd)
	cmd.AddCommand(silenceCmd)

	return cmd
}

func NewAlertmanagerCreateCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create resources in Alertmanager.",
		Long:  "Create resources in Alertmanager for a tenant.",
	}

	// Create silence command.
	var (
		silenceFilePath string
		silenceID       string
		matchers        []string
		startsAt        string
		endsAt          string
		duration        string
		createdBy       string
		comment         string
	)
	silenceCmd := &cobra.Command{
		Use:   "silence",
		Short: "Create or update a silence in Alertmanager.",
		Long: `Create or update a silence in Alertmanager for a tenant.

You can either provide a JSON file with the silence configuration using --file,
or use flags to build the silence configuration.

Matcher formats:
  name=value      Exact match
  name!=value     Negative exact match
  name=~value     Regex match
  name!~value     Negative regex match

Examples:
  # Create a silence for all critical alerts for 2 hours
  obsctl alertmanager create silence \
    --matcher alertname=HighCPU \
    --matcher severity=critical \
    --created-by "John Doe" \
    --comment "Maintenance window" \
    --duration 2h

  # Create a silence with regex matcher
  obsctl alertmanager create silence \
    --matcher "alertname=~(HighCPU|HighMemory)" \
    --created-by "Jane Doe" \
    --comment "Planned maintenance" \
    --starts-at "2024-01-01T10:00:00Z" \
    --ends-at "2024-01-01T12:00:00Z"

  # Update an existing silence (requires --id)
  obsctl alertmanager create silence \
    --id abc-123-def \
    --matcher alertname=HighCPU \
    --created-by "John Doe" \
    --comment "Extended maintenance" \
    --duration 4h

  # Create from JSON file
  obsctl alertmanager create silence --file silence.json`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			var silenceBody models.PostableSilence

			if silenceFilePath != "" {
				// Read silence from file
				file, err := os.Open(silenceFilePath)
				if err != nil {
					return fmt.Errorf("opening silence file: %w", err)
				}
				defer file.Close()

				if err := json.NewDecoder(file).Decode(&silenceBody); err != nil {
					return fmt.Errorf("decoding silence file: %w", err)
				}
			} else {
				// Build silence from flags
				if len(matchers) == 0 {
					return fmt.Errorf("at least one matcher is required")
				}
				if comment == "" {
					return fmt.Errorf("comment is required")
				}
				if createdBy == "" {
					return fmt.Errorf("created-by is required")
				}

				// Parse matchers using alertmanager's label parser
				matcherObjs := make([]models.Matcher, 0, len(matchers))
				for _, m := range matchers {
					parsed, err := labels.ParseMatcher(m)
					if err != nil {
						return fmt.Errorf("invalid matcher %q: %w", m, err)
					}

					isEqual := parsed.Type == labels.MatchEqual || parsed.Type == labels.MatchRegexp
					isRegex := parsed.Type == labels.MatchRegexp || parsed.Type == labels.MatchNotRegexp

					matcherObjs = append(matcherObjs, models.Matcher{
						Name:    parsed.Name,
						Value:   parsed.Value,
						IsEqual: &isEqual,
						IsRegex: isRegex,
					})
				}

				// Calculate time
				var start, end time.Time
				now := time.Now()

				if startsAt != "" {
					start, err = time.Parse(time.RFC3339, startsAt)
					if err != nil {
						return fmt.Errorf("parsing starts-at: %w", err)
					}
				} else {
					start = now
				}

				if endsAt != "" {
					end, err = time.Parse(time.RFC3339, endsAt)
					if err != nil {
						return fmt.Errorf("parsing ends-at: %w", err)
					}
				} else if duration != "" {
					d, err := time.ParseDuration(duration)
					if err != nil {
						return fmt.Errorf("parsing duration: %w", err)
					}
					end = start.Add(d)
				} else {
					// Default to 2 hours
					end = start.Add(2 * time.Hour)
				}

				silenceBody = models.PostableSilence{
					Matchers:  matcherObjs,
					StartsAt:  start,
					EndsAt:    end,
					CreatedBy: createdBy,
					Comment:   comment,
				}

				// Add ID if provided (for updates)
				if silenceID != "" {
					silenceBody.Id = &silenceID
				}
			}

			params := &client.PostSilenceParams{}
			resp, err := f.PostSilenceWithResponse(ctx, currentTenant, params, silenceBody)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			if resp.StatusCode()/100 != 2 {
				if len(resp.Body) != 0 {
					fmt.Fprintln(cmd.OutOrStdout(), string(resp.Body))
					return fmt.Errorf("request failed with status code %d", resp.StatusCode())
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(resp.Body))
			return nil
		},
	}

	silenceCmd.Flags().StringVar(&silenceFilePath, "file", "", "Path to JSON file containing silence configuration")
	silenceCmd.Flags().StringVar(&silenceID, "id", "", "Silence ID (for updating an existing silence)")
	silenceCmd.Flags().StringArrayVarP(&matchers, "matcher", "m", nil, "Matchers for the silence. Supported formats: name=value, name!=value, name=~regex, name!~regex")
	silenceCmd.Flags().StringVar(&startsAt, "starts-at", "", "Start time for the silence (RFC3339 format, e.g., 2024-01-01T10:00:00Z). Defaults to now.")
	silenceCmd.Flags().StringVar(&endsAt, "ends-at", "", "End time for the silence (RFC3339 format, e.g., 2024-01-01T12:00:00Z)")
	silenceCmd.Flags().StringVar(&duration, "duration", "", "Duration for the silence (e.g., 2h, 30m, 1h30m). Defaults to 2h if --ends-at not provided.")
	silenceCmd.Flags().StringVar(&createdBy, "created-by", "", "Creator of the silence (required)")
	silenceCmd.Flags().StringVar(&comment, "comment", "", "Comment describing the silence (required)")

	cmd.AddCommand(silenceCmd)

	return cmd
}

func NewAlertmanagerDeleteCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources in Alertmanager.",
		Long:  "Delete resources in Alertmanager for a tenant.",
	}

	// Delete silence command.
	deleteSilenceCmd := &cobra.Command{
		Use:          "silence [silenceID]",
		Short:        "Delete a silence in Alertmanager.",
		Long:         "Delete a silence in Alertmanager by its ID.",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			silenceID, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid silence ID: %w", err)
			}
			resp, err := f.DeleteSilenceWithResponse(ctx, currentTenant, silenceID)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			if resp.StatusCode()/100 != 2 {
				if len(resp.Body) != 0 {
					fmt.Fprintln(cmd.OutOrStdout(), string(resp.Body))
					return fmt.Errorf("request failed with status code %d", resp.StatusCode())
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Silence %s deleted successfully\n", args[0])
			return nil
		},
	}

	cmd.AddCommand(deleteSilenceCmd)

	return cmd
}

func NewAlertmanagerCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alertmanager",
		Short: "Alertmanager operations for Observatorium.",
		Long:  "Alertmanager operations for Observatorium.",
	}

	cmd.AddCommand(NewAlertmanagerGetCmd(ctx))
	cmd.AddCommand(NewAlertmanagerCreateCmd(ctx))
	cmd.AddCommand(NewAlertmanagerDeleteCmd(ctx))

	return cmd
}

// renderAlertsTable renders alerts in a table format.
func renderAlertsTable(w io.Writer, alerts models.GettableAlerts) {
	table := tablewriter.NewTable(w,
		tablewriter.WithAlignment(tw.Alignment{tw.AlignLeft, tw.AlignLeft, tw.AlignLeft, tw.AlignLeft, tw.AlignLeft}),
		tablewriter.WithRowAutoWrap(0),
		tablewriter.WithRendition(tw.Rendition{Borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off}}),
	)
	table.Header("ALERTNAME", "STATE", "STARTED", "LABELS", "RECEIVERS")

	for _, alert := range alerts {
		alertname := alert.Labels["alertname"]
		state := string(alert.Status.State)

		// Format started time as relative duration
		started := formatDuration(time.Since(alert.StartsAt))

		// Format labels (excluding alertname)
		labelParts := formatLabels(alert.Labels, "alertname")

		// Format receivers
		var receiverNames []string
		for _, r := range alert.Receivers {
			receiverNames = append(receiverNames, r.Name)
		}
		receivers := strings.Join(receiverNames, ",")

		table.Append([]string{alertname, state, started, labelParts, receivers})
	}

	table.Render()
}

// renderSilencesTable renders silences in a table format.
func renderSilencesTable(w io.Writer, silences models.GettableSilences) {
	table := tablewriter.NewTable(w,
		tablewriter.WithAlignment(tw.Alignment{tw.AlignLeft, tw.AlignLeft, tw.AlignLeft, tw.AlignLeft, tw.AlignLeft, tw.AlignLeft, tw.AlignLeft}),
		tablewriter.WithRowAutoWrap(0),
		tablewriter.WithRendition(tw.Rendition{Borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off}}),
	)
	table.Header("ID", "STATE", "MATCHERS", "STARTS", "ENDS", "CREATED BY", "COMMENT")

	for _, silence := range silences {
		state := string(silence.Status.State)

		// Format matchers
		matchers := formatMatchers(silence.Matchers)

		// Format times
		starts := silence.StartsAt.Format(time.RFC3339)
		ends := silence.EndsAt.Format(time.RFC3339)

		// Truncate comment if too long
		comment := silence.Comment
		if len(comment) > 40 {
			comment = comment[:37] + "..."
		}

		// Truncate ID for display (first 8 chars)
		id := silence.Id
		if len(id) > 8 {
			id = id[:8]
		}

		table.Append([]string{id, state, matchers, starts, ends, silence.CreatedBy, comment})
	}

	table.Render()
}

// renderSilenceTable renders a single silence in a table format.
func renderSilenceTable(w io.Writer, silence models.GettableSilence) {
	table := tablewriter.NewTable(w,
		tablewriter.WithAlignment(tw.Alignment{tw.AlignLeft, tw.AlignLeft}),
		tablewriter.WithRowAutoWrap(0),
		tablewriter.WithRendition(tw.Rendition{Borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off}}),
	)

	table.Append([]string{"ID:", silence.Id})
	table.Append([]string{"State:", string(silence.Status.State)})
	table.Append([]string{"Matchers:", formatMatchers(silence.Matchers)})
	table.Append([]string{"Starts At:", silence.StartsAt.Format(time.RFC3339)})
	table.Append([]string{"Ends At:", silence.EndsAt.Format(time.RFC3339)})
	table.Append([]string{"Created By:", silence.CreatedBy})
	table.Append([]string{"Comment:", silence.Comment})
	table.Append([]string{"Updated At:", silence.UpdatedAt.Format(time.RFC3339)})

	table.Render()
}

// formatLabels formats a LabelSet as a comma-separated string, excluding specified keys.
func formatLabels(labels models.LabelSet, exclude ...string) string {
	excludeMap := make(map[string]bool)
	for _, e := range exclude {
		excludeMap[e] = true
	}

	var parts []string
	// Sort keys for consistent output
	keys := make([]string, 0, len(labels))
	for k := range labels {
		if !excludeMap[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, labels[k]))
	}
	return strings.Join(parts, ",")
}

// formatMatchers formats matchers as a comma-separated string.
func formatMatchers(matchers models.Matchers) string {
	var parts []string
	for _, m := range matchers {
		op := "="
		if m.IsRegex {
			if m.IsEqual != nil && !*m.IsEqual {
				op = "!~"
			} else {
				op = "=~"
			}
		} else if m.IsEqual != nil && !*m.IsEqual {
			op = "!="
		}
		parts = append(parts, fmt.Sprintf("%s%s%s", m.Name, op, m.Value))
	}
	return strings.Join(parts, ",")
}

// formatDuration formats a duration in a human-readable short format.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
