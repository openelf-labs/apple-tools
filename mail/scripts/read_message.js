ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var mail = Application("Mail");
mail.includeStandardAdditions = true;

var targetId = params.messageId;
var maxLength = params.maxLength || 10000;

// Recursively search mailboxes for a message by ID
function findMessageInMailboxes(mailboxes) {
  for (var mi = 0; mi < mailboxes.length; mi++) {
    try {
      var found = mailboxes[mi].messages.whose({
        messageId: { _equals: targetId },
      })();
      if (found.length > 0) {
        return found[0];
      }
    } catch (e) {}

    // Recurse into child mailboxes
    try {
      var children = mailboxes[mi].mailboxes();
      if (children.length > 0) {
        var result = findMessageInMailboxes(children);
        if (result) return result;
      }
    } catch (e) {}
  }
  return null;
}

var msg = null;
var accounts = mail.accounts();

for (var ai = 0; ai < accounts.length; ai++) {
  msg = findMessageInMailboxes(accounts[ai].mailboxes());
  if (msg) break;
}

if (!msg) {
  throw new Error("Message not found: " + targetId);
}

var body = "";
try {
  body = msg.content() || "";
} catch (e) {
  body = "(unable to read message body)";
}

// Truncate body
if (body.length > maxLength) {
  body = body.substring(0, maxLength) + "\n...[truncated at " + maxLength + " characters]";
}

// Collect recipients
var toRecipients = [];
try {
  var tos = msg.toRecipients();
  for (var i = 0; i < tos.length; i++) {
    toRecipients.push({
      name: tos[i].name() || "",
      address: tos[i].address(),
    });
  }
} catch (e) {}

var ccRecipients = [];
try {
  var ccs = msg.ccRecipients();
  for (var i = 0; i < ccs.length; i++) {
    ccRecipients.push({
      name: ccs[i].name() || "",
      address: ccs[i].address(),
    });
  }
} catch (e) {}

JSON.stringify({
  messageId: msg.messageId(),
  subject: msg.subject() || "(no subject)",
  sender: msg.sender(),
  dateReceived: msg.dateReceived().toISOString(),
  body: body,
  recipients: {
    to: toRecipients,
    cc: ccRecipients,
  },
});
