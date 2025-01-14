package decision

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"sync"
	"time"

	"github.com/phroggyy/decision/pkg/metadata"
	"github.com/phroggyy/decision/pkg/provider"

	"github.com/gosimple/slug"

	"github.com/slack-go/slack"
)

var (
	categoryOptions []*slack.OptionBlockObject
	categoryLock    sync.Mutex
)

type Client struct {
	api         *slack.Client
	gitProvider provider.Provider
}

func NewClient(token string, gitProvider provider.Provider) *Client {
	return &Client{
		api:         slack.New(token),
		gitProvider: gitProvider,
	}
}

func (c *Client) GetAPI() *slack.Client {
	return c.api
}

func (c *Client) OpenDecisionModal(triggerID string, triggerChannel string, options ...Option) (*slack.View, error) {
	defaults := &Decision{}

	for _, option := range options {
		option(defaults)
	}

	titleLabel := slack.NewTextBlockObject(slack.PlainTextType, "Title", false, false)
	titlePlaceholderText := slack.NewTextBlockObject(slack.PlainTextType, "Give this decision a tl;dr title", false, false)
	titleInput := slack.NewPlainTextInputBlockElement(nil, TitleInputID)
	titleInput.MaxLength = 60
	titleInput.InitialValue = defaults.Title
	titleSection := slack.NewInputBlock(TitleBlockID, titleLabel, titlePlaceholderText, titleInput)

	categoryLabel := slack.NewTextBlockObject(slack.PlainTextType, "Category", false, false)
	categorySelect := slack.NewOptionsSelectBlockElement(slack.OptTypeExternal, nil, CategorySelectID)
	categorySelect.MinQueryLength = new(int)
	categorySection := slack.NewInputBlock(CategoryBlockID, categoryLabel, nil, categorySelect)

	contextLabel := slack.NewTextBlockObject(slack.PlainTextType, "Context", false, false)
	contextPlaceholderText := slack.NewTextBlockObject(slack.PlainTextType, "Explain why this decision needs to be made. What forces are at play?", false, false)
	contextInput := slack.NewPlainTextInputBlockElement(nil, ContextInputID)
	contextInput.Multiline = true
	contextInput.InitialValue = defaults.Context
	contextSection := slack.NewInputBlock(ContextBlockID, contextLabel, contextPlaceholderText, contextInput)
	contextSection.Optional = true

	decisionLabel := slack.NewTextBlockObject(slack.PlainTextType, "Decision", false, false)
	decisionPlaceholderText := slack.NewTextBlockObject(slack.PlainTextType, "Document the decision you've made. Use active voice: \"We will...\"", false, false)
	decisionInput := slack.NewPlainTextInputBlockElement(nil, DecisionInputID)
	decisionInput.Multiline = true
	decisionInput.InitialValue = defaults.Decision
	decisionSection := slack.NewInputBlock(DecisionBlockID, decisionLabel, decisionPlaceholderText, decisionInput)
	decisionSection.Optional = true

	consequencesLabel := slack.NewTextBlockObject(slack.PlainTextType, "Consequences", false, false)
	consequencesPlaceholderText := slack.NewTextBlockObject(slack.PlainTextType, "Describe the consequences, good and bad, after this decision has been made.", false, false)
	consequencesInput := slack.NewPlainTextInputBlockElement(nil, ConsequencesInputID)
	consequencesInput.Multiline = true
	consequencesInput.InitialValue = defaults.Consequences
	consequencesSection := slack.NewInputBlock(ConsequencesBlockID, consequencesLabel, consequencesPlaceholderText, consequencesInput)
	consequencesSection.Optional = true

	view := slack.ModalViewRequest{
		CallbackID:      LogDecisionCallbackID,
		Type:            slack.ViewType("modal"),
		Title:           slack.NewTextBlockObject(slack.PlainTextType, "Record a decision", false, false),
		Close:           slack.NewTextBlockObject(slack.PlainTextType, "Cancel", false, false),
		Submit:          slack.NewTextBlockObject(slack.PlainTextType, "Record decision", false, false),
		PrivateMetadata: metadata.ForChannel(triggerChannel).String(),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				titleSection,
				categorySection,
				contextSection,
				decisionSection,
				consequencesSection,
			},
		},
	}

	viewRes, err := c.api.OpenView(triggerID, view)
	if err != nil {
		fmt.Printf("Error opening modal view: %v\n", err)
	}

	return &viewRes.View, err
}

func (c *Client) GetCategoryOptions(typeAheadValue *string) slack.OptionsResponse {
	// we only fetch the existing categories from github when the modal is first show
	// subsequent calls (sent as the user is typing) re-use the list from this fetch
	modalFirstOpened := typeAheadValue != nil && *typeAheadValue == ""
	if modalFirstOpened {
		categoryLock.Lock()
		defer categoryLock.Unlock()

		categoryOptions = make([]*slack.OptionBlockObject, 0)
		existingFolders, _ := c.gitProvider.GetFolders()

		//existingFolders, _ := github.GetFolders()
		//existingFolders, _ := github.GetFolders()
		for _, folder := range existingFolders {
			categoryOptions = append(categoryOptions, slack.NewOptionBlockObject(
				strings.ToLower(folder),
				slack.NewTextBlockObject(
					slack.PlainTextType, folder, false, false,
				),
				nil,
			))
		}
	}

	// add the type ahead value as an option in the list
	var responseOptions = categoryOptions
	if typeAheadValue != nil && *typeAheadValue != "" {
		typeAheadOption := slack.NewOptionBlockObject(
			slug.Make(*typeAheadValue),
			slack.NewTextBlockObject(slack.PlainTextType, *typeAheadValue+" (Create new)", false, false),
			nil,
		)

		// only add it if it doesn't exist already
		if typeAheadOption != nil {
			typeAheadExists := false
			for _, category := range categoryOptions {
				if typeAheadOption.Value == category.Value {
					typeAheadExists = true
					break
				}
			}

			if !typeAheadExists {
				responseOptions = append([]*slack.OptionBlockObject{typeAheadOption}, categoryOptions...)
			}
		}
	}

	response := slack.OptionsResponse{
		Options: responseOptions,
	}

	return response
}

func (c *Client) HandleModalSubmission(payload *slack.InteractionCallback) error {
	submissionValues := payload.View.State.Values

	sourceChannel := metadata.MustParse(payload.View.PrivateMetadata).ChannelID

	title := submissionValues[TitleBlockID][TitleInputID].Value
	category := submissionValues[CategoryBlockID][CategorySelectID].SelectedOption.Value
	context := submissionValues[ContextBlockID][ContextInputID].Value
	decision := submissionValues[DecisionBlockID][DecisionInputID].Value
	consequences := submissionValues[ConsequencesBlockID][ConsequencesInputID].Value

	username := payload.User.Name
	if payload.User.Profile.DisplayName != "" {
		username = payload.User.Profile.DisplayName
	}

	decisionData := Decision{
		Title:        title,
		SlackHandle:  username,
		TeamID:       payload.Team.ID,
		UserID:       payload.User.ID,
		Category:     category,
		Date:         time.Now().Format("2006-01-02"),
		Context:      context,
		Decision:     decision,
		Consequences: consequences,
	}

	tmpl, err := template.New("decision").Parse(decisionTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template: %v", err)
		return err
	}

	var decisionBytes bytes.Buffer
	err = tmpl.Execute(&decisionBytes, decisionData)
	if err != nil {
		fmt.Printf("Failed to execute template: %v", err)
		return err
	}

	dateNow := time.Now().Format("2006-01-02")
	fileName := category + "/" + dateNow + "-" + slug.Make(title) + ".md"
	commitMessage := title
	content := decisionBytes.Bytes()

	if CommitAsPRs {
		prURL, err := c.gitProvider.RaisePullRequest(slug.Make(title), commitMessage, fileName, content)
		if err != nil {
			return err
		}

		message := "✅ A pull request for \"" + title + "\" has been created <" + prURL + "|here>."
		c.sendDecisionLinkToUser(message, title, prURL, sourceChannel, payload.User.ID)
	} else {
		decisionURL, err := c.gitProvider.CreateCommit(commitMessage, fileName, content)

		if err != nil {
			return err
		}

		message := "✅ Your decision \"" + title + "\" has been committed <" + decisionURL + "|here>."
		c.sendDecisionLinkToUser(message, title, decisionURL, sourceChannel, payload.User.ID)
	}

	return nil
}

func (c *Client) sendDecisionLinkToUser(message string, title string, fileURL string, channel string, user string) {
	// Return an ephemeral message to the user
	msgOption := slack.MsgOptionText(message, false)
	_, err := c.api.PostEphemeral(channel, user, msgOption)
	if err != nil {
		fmt.Printf("Failed to send message: %v (%v, %v)\n", err, channel, user)
		return
	}
}
