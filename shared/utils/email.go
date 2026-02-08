package utils

import (
	"context"
	"encoding/json"
	"fmt"

	appconfig "shared/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog"
)

/* ───────────────────────── Types ───────────────────────── */

type EmailService struct {
	sqsClient *sqs.Client
	queueURL  string
	fromEmail string
	logger    zerolog.Logger
}

type EmailMessage struct {
	To           []string          `json:"to"`
	From         string            `json:"from"`
	Subject      string            `json:"subject"`
	TemplateName string            `json:"template_name"`
	TemplateData map[string]string `json:"template_data"`
}

/* ───────────────────────── Constructor ───────────────────────── */

func NewEmailService(
	cfg *appconfig.AWSConfig,
	logger zerolog.Logger,
) (*EmailService, error) {

	ctx := context.Background()

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.AWSRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AWSAccessKey,
				cfg.AWSSecretKey,
				"",
			),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("aws config load failed: %w", err)
	}

	logger.Info().
		Str("region", cfg.AWSRegion).
		Msg("AWS SQS initialized")

	return &EmailService{
		sqsClient: sqs.NewFromConfig(awsCfg),
		queueURL:  cfg.AWSSQSQueueURL,
		fromEmail: cfg.AWSFromEmail,
		logger:    logger,
	}, nil
}

/* ───────────────────────── Internal ───────────────────────── */

func (s *EmailService) send(ctx context.Context, msg EmailMessage) error {

	if len(msg.To) == 0 {
		return fmt.Errorf("recipient is empty")
	}

	if msg.From == "" {
		msg.From = s.fromEmail
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	out, err := s.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueURL),
		MessageBody: aws.String(string(body)),
	})

	if err != nil {
		s.logger.Error().
			Err(err).
			Str("to", msg.To[0]).
			Msg("SQS send failed")

		return fmt.Errorf("sqs send: %w", err)
	}

	s.logger.Info().
		Str("id", *out.MessageId).
		Str("to", msg.To[0]).
		Str("template", msg.TemplateName).
		Msg("email queued")

	return nil
}

/* ───────────────────────── Public API ───────────────────────── */

func (s *EmailService) SendVerification(
	ctx context.Context,
	to, token, verifyURL, name string,
) error {

	msg := EmailMessage{
		To:           []string{to},
		Subject:      "Verify Your Email",
		TemplateName: "VerifyEmail",
		TemplateData: map[string]string{
			"url":  fmt.Sprintf("%s?token=%s", verifyURL, token),
			"name": name,
		},
	}

	return s.send(ctx, msg)
}

func (s *EmailService) SendPasswordReset(
	ctx context.Context,
	to, token, resetURL, name string,
) error {

	msg := EmailMessage{
		To:           []string{to},
		Subject:      "Reset Your Password",
		TemplateName: "ResetPassword",
		TemplateData: map[string]string{
			"url":  fmt.Sprintf("%s?token=%s", resetURL, token),
			"name": name,
		},
	}

	return s.send(ctx, msg)
}