package utils

import (
	"context"
	"encoding/json"
	"fmt"

	appConfig "authentication/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config" 
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog"
)

/* ───────────────────────── Types ─────────────────────────────────────── */

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

/* ───────────────────────── Constructor ───────────────────────────────── */

// FIX 2: Parameter type ko 'appConfig.AWSConfig' kar diya
func NewEmailService(
	cfg *appConfig.AWSConfig, 
	logger zerolog.Logger,
) (*EmailService, error) {
	ctx := context.Background()
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.AWSRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AWSAccessKey, cfg.AWSSecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	logger.Info().Str("region", cfg.AWSRegion).Msg("AWS SQS client initialized")

	return &EmailService{
		sqsClient: sqs.NewFromConfig(awsCfg),
		queueURL:  cfg.AWSSQSQueueURL, // Correct field name from config struct
		fromEmail: cfg.AWSFromEmail,   // Correct field name from config struct
		logger:    logger,
	}, nil
}

/* ───────────────────────── Internal ──────────────────────────────────── */

func (s *EmailService) sendEmail(ctx context.Context, msg EmailMessage) error {
	if msg.From == "" {
		msg.From = s.fromEmail
	}
	if len(msg.To) == 0 {
		return fmt.Errorf("sendEmail: To field is empty")
	}

	body, err := json.Marshal(msg)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to marshal email message")
		return fmt.Errorf("marshal message: %w", err)
	}

	result, err := s.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueURL),
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		// FIX 4: Safe check to ensure msg.To exists before logging
		toEmail := ""
		if len(msg.To) > 0 {
			toEmail = msg.To[0]
		}
		s.logger.Error().Err(err).Str("to", toEmail).Msg("failed to send message to SQS")
		return fmt.Errorf("send to SQS: %w", err)
	}

	s.logger.Info().
		Str("message_id", *result.MessageId).
		Str("to", msg.To[0]).
		Str("template", msg.TemplateName).
		Msg("email queued successfully")

	return nil
}

/* ───────────────────────── Public API ────────────────────────────────── */
func (s *EmailService) SendVerificationEmail(ctx context.Context, to, token, verifyURL, name string) error {
	msg := EmailMessage{
		To:           []string{to},
		From:         s.fromEmail,
		Subject:      "Verify Your Email Address",
		TemplateName: "VerifyEmail",
		TemplateData: map[string]string{
			"verification_url": fmt.Sprintf("%s?token=%s", verifyURL, token),
			"user_name":        name,
		},
	}

	if err := s.sendEmail(ctx, msg); err != nil {
		s.logger.Error().Err(err).Str("email", to).Msg("failed to send verification email")
		return fmt.Errorf("send verification email: %w", err)
	}
	return nil
}

func (s *EmailService) SendPasswordResetEmail(ctx context.Context, to, token, resetURL, name string) error {
	msg := EmailMessage{
		To:           []string{to},
		From:         s.fromEmail,
		Subject:      "Reset Your Password",
		TemplateName: "ResetPassword",
		TemplateData: map[string]string{
			"reset_url": fmt.Sprintf("%s?token=%s", resetURL, token),
			"user_name": name,
		},
	}

	if err := s.sendEmail(ctx, msg); err != nil {
		s.logger.Error().Err(err).Str("email", to).Msg("failed to send password reset email")
		return fmt.Errorf("send password reset email: %w", err)
	}
	return nil
}



