package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	tplHtml "html/template"
	"log/slog"
	"reflect"
	"strings"
	tplText "text/template"
	"time"

	dbmodels "github.com/ringecosystem/degov-apps/database/models"
	gqlmodels "github.com/ringecosystem/degov-apps/graph/models"
	"github.com/ringecosystem/degov-apps/internal"
	"github.com/ringecosystem/degov-apps/internal/config"
	"github.com/ringecosystem/degov-apps/internal/templates"
	"github.com/ringecosystem/degov-apps/internal/utils"
	"github.com/ringecosystem/degov-apps/types"
)

type TemplateService struct {
	daoService       *DaoService
	proposalService  *ProposalService
	daoConfigService *DaoConfigService
	htmlTemplates    *tplHtml.Template
	textTemplates    *tplText.Template
}

func NewTemplateService() *TemplateService {
	htmlTmpls := tplHtml.Must(tplHtml.New("").ParseFS(
		templates.TemplateFS,
		"template/*.html",
	))

	textTmpls := tplText.Must(tplText.New("").ParseFS(
		templates.TemplateFS,
		"template/*.md",
	))
	return &TemplateService{
		daoService:       NewDaoService(),
		proposalService:  NewProposalService(),
		daoConfigService: NewDaoConfigService(),
		htmlTemplates:    htmlTmpls,
		textTemplates:    textTmpls,
	}
}

type TemplateNotificationRecordData struct {
	DegovSiteConfig types.DegovSiteConfig      `json:"degov_site_config"`
	DaoConfig       *types.DaoConfig           `json:"dao_config"`
	Dao             *gqlmodels.Dao             `json:"dao"`
	Proposal        *dbmodels.ProposalTracking `json:"proposal"`
	Vote            *internal.VoteCast         `json:"vote,omitempty"`
	PayloadData     map[string]interface{}     `json:"payload_data"`
	EventID         string                     `json:"event_id"`
	UserID          string                     `json:"user_id"`
	UserAddress     string                     `json:"user_address"`
}

// parsePayload attempts to parse the payload as JSON, falls back to string if failed
func (s *TemplateService) parsePayload(payload *string) map[string]interface{} {
	result := make(map[string]interface{})

	if payload == nil || *payload == "" {
		return result
	}

	// Try to parse as JSON first
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(*payload), &jsonData); err == nil {
		return jsonData
	}

	// If JSON parsing fails, treat as plain string
	result["__raw"] = *payload
	return result
}

func (s *TemplateService) getTemplateFileName(notificationType dbmodels.SubscribeFeatureName, mode string) string {
	switch notificationType {
	case dbmodels.SubscribeFeatureProposalNew:
		return "proposal_new." + mode
	case dbmodels.SubscribeFeatureProposalStateChanged:
		return "proposal_state_changed." + mode
	case dbmodels.SubscribeFeatureVoteEnd:
		return "vote_end." + mode
	case dbmodels.SubscribeFeatureVoteEmitted:
		return "vote_emitted." + mode
	default:
		return "unknown." + mode // fallback
	}
}

func (s *TemplateService) getTemplateTitle(notificationType dbmodels.SubscribeFeatureName) string {
	switch notificationType {
	case dbmodels.SubscribeFeatureProposalNew:
		return "New Proposal"
	case dbmodels.SubscribeFeatureProposalStateChanged:
		return "Proposal State Changed"
	case dbmodels.SubscribeFeatureVoteEnd:
		return "Vote End Reminder"
	case dbmodels.SubscribeFeatureVoteEmitted:
		return "Vote Emitted"
	default:
		return "Unknown" // fallback
	}
}

func (s *TemplateService) GenerateTemplateByNotificationRecord(record *dbmodels.NotificationRecord) (*types.TemplateOutput, error) {
	// Get DAO information
	dao, err := s.daoService.Inspect(types.BasicInput[string]{
		User:  nil,
		Input: record.DaoCode,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get DAO info: %w", err)
	}

	// Get proposal information
	proposal, err := s.proposalService.InspectProposal(types.InpspectProposalInput{
		DaoCode:    record.DaoCode,
		ProposalID: record.ProposalID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal info: %w", err)
	}

	daoConfig, err := s.daoConfigService.StandardConfig(dao.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to get DAO config info: %w", err)
	}

	var vote *internal.VoteCast
	if record.VoteID != nil {
		degovIndexer := internal.NewDegovIndexer(daoConfig.Indexer.Endpoint)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		voteById, err := degovIndexer.QueryVote(ctx, *record.VoteID)
		cancel()

		if err != nil {
			return nil, fmt.Errorf("failed to get vote info: %w", err)
		}

		vote = voteById
	}

	// Parse payload data
	payloadData := s.parsePayload(record.Payload)

	templateData := TemplateNotificationRecordData{
		DegovSiteConfig: config.GetDegovSiteConfig(),
		DaoConfig:       daoConfig,
		Dao:             dao,
		Proposal:        proposal,
		Vote:            vote,
		PayloadData:     payloadData,
		EventID:         record.EventID,
		UserID:          record.UserID,
		UserAddress:     record.UserAddress,
	}

	richTemplateFileName := s.getTemplateFileName(record.Type, "html")
	plainTemplateFileName := s.getTemplateFileName(record.Type, "md")

	richText, err := s.renderTemplate(richTemplateFileName, templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to render rich text template %s: %w", richTemplateFileName, err)
	}
	palinText, err := s.renderTemplate(plainTemplateFileName, templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to render plain text template %s: %w", plainTemplateFileName, err)
	}

	title := fmt.Sprintf("[DeGov] [%s] [%s]: %s", dao.Name, s.getTemplateTitle(record.Type), proposal.Title)

	return &types.TemplateOutput{
		Title:            utils.TruncateText(title, 80),
		RichTextContent:  richText,
		PlainTextContent: palinText,
	}, nil
}

func (s *TemplateService) renderTemplate(templateName string, data interface{}) (string, error) {

	var finData interface{}
	templateData, serr := structToMap(data)
	if serr != nil {
		if m, ok := data.(map[string]interface{}); ok {
			finData = m
		} else {
			finData = data
		}
	} else {
		finData = templateData
	}

	var buf bytes.Buffer
	var err error

	if strings.HasSuffix(templateName, ".html") {
		err = s.htmlTemplates.ExecuteTemplate(&buf, templateName, finData)
	} else if strings.HasSuffix(templateName, ".md") {
		err = s.textTemplates.ExecuteTemplate(&buf, templateName, finData)
	} else {
		return "", fmt.Errorf("unsupported template type: %s", templateName)
	}

	if err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

func (s *TemplateService) GenerateTemplateOTP(input types.GenerateTemplateOTPInput) (*types.TemplateOutput, error) {
	richText, err := s.renderTemplate("otp.html", input)
	if err != nil {
		slog.Error("failed to render OTP html template", "err", err)
		return nil, err
	}
	plainText, err := s.renderTemplate("otp.md", input)
	if err != nil {
		slog.Error("failed to render OTP md template", "err", err)
		return nil, err
	}

	return &types.TemplateOutput{
		Title:            "[DeGov] Email Verification",
		RichTextContent:  richText,
		PlainTextContent: plainText,
	}, nil

}

func structToMap(data interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	val := reflect.ValueOf(data)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("data is not a struct, but a %s", val.Kind())
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.IsExported() {
			result[field.Name] = val.Field(i).Interface()
		}
	}

	return result, nil
}
