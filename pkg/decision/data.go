package decision

var (
	Token       string
	CommitAsPRs bool
)

const (
	SlashCommand = "/decision"

	TitleBlockID = "title_block"
	TitleInputID = "title_input"

	CategoryBlockID  = "category_block"
	CategorySelectID = "category_select"

	ContextBlockID = "context_block"
	ContextInputID = "context_input"

	DecisionBlockID = "decision_block"
	DecisionInputID = "decision_input"

	ConsequencesBlockID = "consequences_block"
	ConsequencesInputID = "consequences_input"

	LogDecisionCallbackID = "log_decision"
)

type Decision struct {
	Title        string
	SlackHandle  string
	TeamID       string
	UserID       string
	Category     string
	Date         string
	Context      string
	Decision     string
	Consequences string
}

type Option = func(*Decision)

func WithTitle(title string) Option {
	return func(d *Decision) {
		d.Title = title
	}
}

func WithContext(context string) Option {
	return func(d *Decision) {
		d.Context = context
	}
}

func WithConsequences(consequences string) Option {
	return func(d *Decision) {
		d.Consequences = consequences
	}
}

func WithDecision(decision string) Option {
	return func(d *Decision) {
		d.Decision = decision
	}
}

var decisionTemplate = `# {{.Title}}

Author: [@{{.SlackHandle}}](slack://user?team={{.TeamID}}&id={{.UserID}})

Category: ` + "`{{.Category}}`" + `

Date: {{.Date}}

## Context

{{.Context}}

## Decision

{{.Decision}}

## Consequences

{{.Consequences}}
`
