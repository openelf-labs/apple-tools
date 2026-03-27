ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var mail = Application("Mail");
mail.includeStandardAdditions = true;

var toAddresses = params.to || [];
var subject = params.subject || "";
var body = params.body || "";
var ccAddresses = params.cc || [];
var bccAddresses = params.bcc || [];
var shouldSend = params.send || false;

var msg = mail.OutgoingMessage({
  subject: subject,
  content: body,
  visible: !shouldSend,
});

mail.outgoingMessages.push(msg);

// Add To recipients
for (var i = 0; i < toAddresses.length; i++) {
  var r = mail.Recipient({ address: toAddresses[i] });
  msg.toRecipients.push(r);
}

// Add CC recipients
for (var i = 0; i < ccAddresses.length; i++) {
  var r = mail.CcRecipient({ address: ccAddresses[i] });
  msg.ccRecipients.push(r);
}

// Add BCC recipients
for (var i = 0; i < bccAddresses.length; i++) {
  var r = mail.BccRecipient({ address: bccAddresses[i] });
  msg.bccRecipients.push(r);
}

var result;
if (shouldSend) {
  msg.send();
  result = {
    success: true,
    message: "Email sent successfully.",
    sent: true,
  };
} else {
  result = {
    success: true,
    message: "Draft created and opened for review.",
    sent: false,
  };
}

JSON.stringify(result);
