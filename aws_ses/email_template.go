package aws_ses

const (
	DeactivateAccountTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Account Deactivation Notice</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    body { font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 0; }
    .container { max-width: 600px; background-color: #ffffff; margin: 40px auto; padding: 20px; border-radius: 8px; }
    .header { font-size: 24px; font-weight: bold; margin-bottom: 20px; color: #d9534f; }
    .content { font-size: 16px; line-height: 1.5; margin-bottom: 30px; }
    .footer { font-size: 12px; color: #888888; margin-top: 30px; }
    .button {
      display: inline-block;
      padding: 12px 24px;
      font-size: 16px;
      color: white;
      background-color: #d9534f;
      border-radius: 5px;
      text-decoration: none;
      margin-top: 20px;
    }
  </style>
</head>
<body>
  <table width="100%" cellpadding="0" cellspacing="0" role="presentation">
    <tr>
      <td align="center">
        <div class="container">
          <div class="header">Account Deactivation Notice</div>
          <div class="content">
            Hello {{name}},<br><br>
            We wanted to let you know that your account with {{company_name}} has been deactivated.<br><br>
            If you did not request this action or believe this is a mistake, please click the button below to reactivate your account.<br><br>
            <a href="{{reactivate_link}}" class="button">Reactivate Account</a>
			<br><br>
			Your account will be deactivated for 30 days. If you do not reactivate your account within 30 days, it will be permanently deleted.
          </div>
          <div class="content">
            If you have any questions or need assistance, please contact our support team.<br><br>
            Thank you for being a part of {{company_name}}.
          </div>
          <div class="footer">
            &copy; {{year}} {{company_name}}. All rights reserved.<br>
            If you did not request this, please ignore this email or contact support.
          </div>
        </div>
      </td>
    </tr>
  </table>
</body>
</html>`
)

const ReactivateAccountTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Account Reactivation Successful</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    body { font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 0; }
    .container { max-width: 600px; background-color: #ffffff; margin: 40px auto; padding: 20px; border-radius: 8px; }
    .header { font-size: 24px; font-weight: bold; margin-bottom: 20px; color: #5cb85c; }
    .content { font-size: 16px; line-height: 1.5; margin-bottom: 30px; }
    .footer { font-size: 12px; color: #888888; margin-top: 30px; }

  </style>
</head>
<body>
  <table width="100%" cellpadding="0" cellspacing="0" role="presentation">
    <tr>
      <td align="center">
        <div class="container">
          <div class="header">Account Reactivated</div>
          <div class="content">
            Hello {{name}},<br><br>
            Your account with {{company_name}} has been successfully reactivated.<br><br>
            You can now continue using all our services as usual.<br><br>
          </div>
          <div class="content">
            If you have any questions or need assistance, please contact our support team.<br><br>
            Thank you for being a valued member of {{company_name}}.
          </div>
          <div class="footer">
            &copy; {{year}} {{company_name}}. All rights reserved.<br>
            If you did not request this, please ignore this email or contact support.
          </div>
        </div>
      </td>
    </tr>
  </table>
</body>
</html>`

const EInvoiceGeneratedTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>E-Invoice Successfully Generated</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    body { font-family: Arial, sans-serif; background-color: #f5f7fa; margin: 0; padding: 0; }
    .container { max-width: 600px; background-color: #ffffff; margin: 40px auto; padding: 32px 24px; border-radius: 10px; box-shadow: 0 2px 8px rgba(44, 62, 80, 0.08); }
    .header { font-size: 26px; font-weight: bold; margin-bottom: 18px; color: #1a73e8; text-align: center; }
    .content { font-size: 17px; color: #444; margin-bottom: 26px; line-height: 1.6; text-align: center; }
    .invoice-box { margin: 0 auto 22px auto; padding: 20px; background-color: #f0f4fb; border-radius: 8px; border: 1px solid #e3e8f0; width: 80%; }
    .invoice-label { font-size: 15px; color: #888; }
    .invoice-number { font-size: 28px; font-weight: bold; letter-spacing: 2px; color: #1a73e8; }
    .button {
      display: inline-block;
      margin-top: 22px;
      font-size: 16px;
      color: #fff;
      background-color: #1a73e8;
      padding: 12px 32px;
      border-radius: 4px;
      text-decoration: none;
      font-weight: 600;
      transition: background-color 0.2s;
    }
    .button:hover { background-color: #155ab6; }
    .footer { font-size: 13px; color: #b1b3b8; margin-top: 32px; text-align: center; }
    @media (max-width:640px) {
      .container { padding: 16px 6px; }
      .invoice-box { width: 100%; padding: 14px; }
    }
  </style>
</head>
<body>
  <table width="100%" cellpadding="0" cellspacing="0" role="presentation">
    <tr>
      <td align="center">
        <div class="container">
          <div class="header">E-Invoice Generated Successfully</div>
          <div class="content">
            Hello {{name}},<br><br>
            We are pleased to inform you that your e-invoice<br>
            <div class="invoice-box">
              <div class="invoice-label">Invoice Number</div>
              <div class="invoice-number">{{invoice_number}}</div>
            </div>
            has been successfully generated.<br><br>
            You can view, download, and keep this invoice for your records.<br>
            If you have any questions regarding this e-invoice, please contact our support team.
          </div>
          <div style="text-align:center;">
            <a href="{{invoice_url}}" class="button">View E-Invoice</a>
          </div>
          <div class="footer">
            <span style="color:#d0d1d4">This email is auto-generated.</span>
          </div>
        </div>
      </td>
    </tr>
  </table>
</body>
</html>`

const (
	PasswordResetTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Password Reset Request</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    body { font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 0; }
    .container { max-width: 600px; background-color: #ffffff; margin: 40px auto; padding: 20px; border-radius: 8px; }
    .header { font-size: 24px; font-weight: bold; margin-bottom: 20px; }
    .content { font-size: 16px; line-height: 1.5; margin-bottom: 30px; }
    .verification-code { font-size: 32px; font-weight: bold; text-align: center; background-color: #f8f9fa; padding: 20px; border-radius: 8px; border: 2px dashed #dee2e6; margin: 20px 0; letter-spacing: 8px; color: #1a73e8; }
    .footer { font-size: 12px; color: #888888; margin-top: 30px; }
  </style>
</head>
<body>
  <table width="100%" cellpadding="0" cellspacing="0" role="presentation">
    <tr>
      <td align="center">
        <div class="container">
          <div class="header">Password Reset Request</div>
          <div class="content">
            Hello {{name}},<br><br>
            We received a request to reset your password for your account.<br><br>
            Please enter the following verification code to reset your password:
          </div>
          <div class="verification-code">
            {{verification_code}}
          </div>
          <div class="content">
            Enter this 6-digit code in the password reset screen.<br><br>
            <strong>This code will expire in 10 minutes for your security.</strong><br><br>
            If you did not request a password reset, you can safely ignore this email and your password will remain unchanged.
          </div>
          <div class="footer">
            &copy; {{year}} {{company_name}}. All rights reserved.
          </div>
        </div>
      </td>
    </tr>
  </table>
</body>
</html>`
)
