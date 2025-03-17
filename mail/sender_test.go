package mail

import (
	"simplebank/util"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendEmail(t *testing.T) {
    if testing.Short(){
        t.Skip()
    }
	config, err := util.LoadConfig("..")
	require.NoError(t, err)
	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	subject := "Simple bank"
	content := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Bank Transaction Notification</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f4f4f4;
            margin: 0;
            padding: 0;
        }
        .email-container {
            max-width: 600px;
            margin: 20px auto;
            background-color: #ffffff;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
        }
        .header {
            text-align: center;
            padding-bottom: 20px;
            border-bottom: 1px solid #dddddd;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            color: #333333;
        }
        .content {
            padding: 20px 0;
        }
        .content h2 {
            font-size: 20px;
            color: #333333;
            margin-bottom: 10px;
        }
        .content p {
            font-size: 16px;
            color: #555555;
            line-height: 1.6;
        }
        .footer {
            text-align: center;
            padding-top: 20px;
            border-top: 1px solid #dddddd;
            font-size: 14px;
            color: #777777;
        }
    </style>
</head>
<body>
    <div class="email-container">
        <div class="header">
            <h1>Simple Bank</h1>
        </div>
        <div class="content">
            <h2>Transaction Notification</h2>
            <p>Dear Customer,</p>
            <p>We would like to inform you about a recent transaction on your account:</p>
            <p><strong>Transaction Type:</strong> Deposit</p>
            <p><strong>Amount:</strong> 500.00</p>
            <p><strong>Date:</strong> October 10, 2023</p>
            <p><strong>Account Balance:</strong> 1500.00</p>
            <p>If you did not authorize this transaction, please contact us immediately at <a href="mailto:support@simplebank.com">support@simplebank.com</a>.</p>
            <p>Thank you for banking with us!</p>
        </div>
        <div class="footer">
            <p>&copy; 2023 Simple Bank. All rights reserved.</p>
            <p>This is an automated message, please do not reply directly to this email.</p>
        </div>
    </div>
</body>
</html>`

	to := []string{"javakhiryulchibaev@gmail.com"}
	attachedFiles := []string{"../README.md"}
	err = sender.SendEmail(
		subject,
		content,
		to,
		nil,
		nil,
		attachedFiles,
	)
	require.NoError(t, err)
}
