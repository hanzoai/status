package alerting

import (
	"reflect"
	"strings"

	"hanzo.ai/status/alerting/alert"
	"hanzo.ai/status/alerting/provider"
	"hanzo.ai/status/alerting/provider/awsses"
	"hanzo.ai/status/alerting/provider/clickup"
	"hanzo.ai/status/alerting/provider/custom"
	"hanzo.ai/status/alerting/provider/datadog"
	"hanzo.ai/status/alerting/provider/discord"
	"hanzo.ai/status/alerting/provider/email"
	"hanzo.ai/status/alerting/provider/gitea"
	"hanzo.ai/status/alerting/provider/github"
	"hanzo.ai/status/alerting/provider/gitlab"
	"hanzo.ai/status/alerting/provider/googlechat"
	"hanzo.ai/status/alerting/provider/gotify"
	"hanzo.ai/status/alerting/provider/homeassistant"
	"hanzo.ai/status/alerting/provider/ifttt"
	"hanzo.ai/status/alerting/provider/ilert"
	"hanzo.ai/status/alerting/provider/incidentio"
	"hanzo.ai/status/alerting/provider/line"
	"hanzo.ai/status/alerting/provider/matrix"
	"hanzo.ai/status/alerting/provider/mattermost"
	"hanzo.ai/status/alerting/provider/messagebird"
	"hanzo.ai/status/alerting/provider/n8n"
	"hanzo.ai/status/alerting/provider/newrelic"
	"hanzo.ai/status/alerting/provider/ntfy"
	"hanzo.ai/status/alerting/provider/opsgenie"
	"hanzo.ai/status/alerting/provider/pagerduty"
	"hanzo.ai/status/alerting/provider/plivo"
	"hanzo.ai/status/alerting/provider/pushover"
	"hanzo.ai/status/alerting/provider/rocketchat"
	"hanzo.ai/status/alerting/provider/sendgrid"
	"hanzo.ai/status/alerting/provider/signal"
	"hanzo.ai/status/alerting/provider/signl4"
	"hanzo.ai/status/alerting/provider/slack"
	"hanzo.ai/status/alerting/provider/splunk"
	"hanzo.ai/status/alerting/provider/squadcast"
	"hanzo.ai/status/alerting/provider/teams"
	"hanzo.ai/status/alerting/provider/teamsworkflows"
	"hanzo.ai/status/alerting/provider/telegram"
	"hanzo.ai/status/alerting/provider/twilio"
	"hanzo.ai/status/alerting/provider/vonage"
	"hanzo.ai/status/alerting/provider/webex"
	"hanzo.ai/status/alerting/provider/zapier"
	"hanzo.ai/status/alerting/provider/zulip"
	"github.com/TwiN/logr"
)

// Config is the configuration for alerting providers
type Config struct {
	// AWSSimpleEmailService is the configuration for the aws-ses alerting provider
	AWSSimpleEmailService *awsses.AlertProvider `yaml:"aws-ses,omitempty"`

	// ClickUp is the configuration for the clickup alerting provider
	ClickUp *clickup.AlertProvider `yaml:"clickup,omitempty"`

	// Custom is the configuration for the custom alerting provider
	Custom *custom.AlertProvider `yaml:"custom,omitempty"`

	// Datadog is the configuration for the datadog alerting provider
	Datadog *datadog.AlertProvider `yaml:"datadog,omitempty"`

	// Discord is the configuration for the discord alerting provider
	Discord *discord.AlertProvider `yaml:"discord,omitempty"`

	// Email is the configuration for the email alerting provider
	Email *email.AlertProvider `yaml:"email,omitempty"`

	// GitHub is the configuration for the github alerting provider
	GitHub *github.AlertProvider `yaml:"github,omitempty"`

	// GitLab is the configuration for the gitlab alerting provider
	GitLab *gitlab.AlertProvider `yaml:"gitlab,omitempty"`

	// Gitea is the configuration for the gitea alerting provider
	Gitea *gitea.AlertProvider `yaml:"gitea,omitempty"`

	// GoogleChat is the configuration for the googlechat alerting provider
	GoogleChat *googlechat.AlertProvider `yaml:"googlechat,omitempty"`

	// Gotify is the configuration for the gotify alerting provider
	Gotify *gotify.AlertProvider `yaml:"gotify,omitempty"`

	// HomeAssistant is the configuration for the homeassistant alerting provider
	HomeAssistant *homeassistant.AlertProvider `yaml:"homeassistant,omitempty"`

	// IFTTT is the configuration for the ifttt alerting provider
	IFTTT *ifttt.AlertProvider `yaml:"ifttt,omitempty"`

	// Ilert is the configuration for the ilert alerting provider
	Ilert *ilert.AlertProvider `yaml:"ilert,omitempty"`

	// IncidentIO is the configuration for the incident-io alerting provider
	IncidentIO *incidentio.AlertProvider `yaml:"incident-io,omitempty"`

	// Line is the configuration for the line alerting provider
	Line *line.AlertProvider `yaml:"line,omitempty"`

	// Matrix is the configuration for the matrix alerting provider
	Matrix *matrix.AlertProvider `yaml:"matrix,omitempty"`

	// Mattermost is the configuration for the mattermost alerting provider
	Mattermost *mattermost.AlertProvider `yaml:"mattermost,omitempty"`

	// Messagebird is the configuration for the messagebird alerting provider
	Messagebird *messagebird.AlertProvider `yaml:"messagebird,omitempty"`

	// NewRelic is the configuration for the newrelic alerting provider
	NewRelic *newrelic.AlertProvider `yaml:"newrelic,omitempty"`

	// N8N is the configuration for the n8n alerting provider
	N8N *n8n.AlertProvider `yaml:"n8n,omitempty"`

	// Ntfy is the configuration for the ntfy alerting provider
	Ntfy *ntfy.AlertProvider `yaml:"ntfy,omitempty"`

	// Opsgenie is the configuration for the opsgenie alerting provider
	Opsgenie *opsgenie.AlertProvider `yaml:"opsgenie,omitempty"`

	// PagerDuty is the configuration for the pagerduty alerting provider
	PagerDuty *pagerduty.AlertProvider `yaml:"pagerduty,omitempty"`

	// Plivo is the configuration for the plivo alerting provider
	Plivo *plivo.AlertProvider `yaml:"plivo,omitempty"`

	// Pushover is the configuration for the pushover alerting provider
	Pushover *pushover.AlertProvider `yaml:"pushover,omitempty"`

	// RocketChat is the configuration for the rocketchat alerting provider
	RocketChat *rocketchat.AlertProvider `yaml:"rocketchat,omitempty"`

	// SendGrid is the configuration for the sendgrid alerting provider
	SendGrid *sendgrid.AlertProvider `yaml:"sendgrid,omitempty"`

	// Signal is the configuration for the signal alerting provider
	Signal *signal.AlertProvider `yaml:"signal,omitempty"`

	// SIGNL4 is the configuration for the signl4 alerting provider
	SIGNL4 *signl4.AlertProvider `yaml:"signl4,omitempty"`

	// Slack is the configuration for the slack alerting provider
	Slack *slack.AlertProvider `yaml:"slack,omitempty"`

	// Splunk is the configuration for the splunk alerting provider
	Splunk *splunk.AlertProvider `yaml:"splunk,omitempty"`

	// Squadcast is the configuration for the squadcast alerting provider
	Squadcast *squadcast.AlertProvider `yaml:"squadcast,omitempty"`

	// Teams is the configuration for the teams alerting provider
	Teams *teams.AlertProvider `yaml:"teams,omitempty"`

	// TeamsWorkflows is the configuration for the teams alerting provider using the new Workflow App Webhook Connector
	TeamsWorkflows *teamsworkflows.AlertProvider `yaml:"teams-workflows,omitempty"`

	// Telegram is the configuration for the telegram alerting provider
	Telegram *telegram.AlertProvider `yaml:"telegram,omitempty"`

	// Twilio is the configuration for the twilio alerting provider
	Twilio *twilio.AlertProvider `yaml:"twilio,omitempty"`

	// Vonage is the configuration for the vonage alerting provider
	Vonage *vonage.AlertProvider `yaml:"vonage,omitempty"`

	// Webex is the configuration for the webex alerting provider
	Webex *webex.AlertProvider `yaml:"webex,omitempty"`

	// Zapier is the configuration for the zapier alerting provider
	Zapier *zapier.AlertProvider `yaml:"zapier,omitempty"`

	// Zulip is the configuration for the zulip alerting provider
	Zulip *zulip.AlertProvider `yaml:"zulip,omitempty"`
}

// GetAlertingProviderByAlertType returns an provider.AlertProvider by its corresponding alert.Type
func (config *Config) GetAlertingProviderByAlertType(alertType alert.Type) provider.AlertProvider {
	entityType := reflect.TypeOf(config).Elem()
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		tag := strings.Split(field.Tag.Get("yaml"), ",")[0]
		if tag == string(alertType) {
			fieldValue := reflect.ValueOf(config).Elem().Field(i)
			if fieldValue.IsNil() {
				return nil
			}
			return fieldValue.Interface().(provider.AlertProvider)
		}
	}
	logr.Infof("[alerting.GetAlertingProviderByAlertType] No alerting provider found for alert type %s", alertType)
	return nil
}

// SetAlertingProviderToNil Sets an alerting provider to nil to avoid having to revalidate it every time an
// alert of its corresponding type is sent.
func (config *Config) SetAlertingProviderToNil(p provider.AlertProvider) {
	entityType := reflect.TypeOf(config).Elem()
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		if field.Type == reflect.TypeOf(p) {
			reflect.ValueOf(config).Elem().Field(i).Set(reflect.Zero(field.Type))
		}
	}
}
