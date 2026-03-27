ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var mail = Application("Mail");
mail.includeStandardAdditions = true;

var targetId = params.messageId;
var replyBody = params.body;
var replyAll = params.replyAll || false;
var shouldSend = params.send || false;

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

// Build reply
var originalSender = msg.sender();
var originalSubject = msg.subject() || "";
var reSubject = originalSubject;
if (reSubject.substring(0, 3).toLowerCase() !== "re:") {
  reSubject = "Re: " + reSubject;
}

var reply = mail.OutgoingMessage({
  subject: reSubject,
  content: replyBody,
  visible: !shouldSend,
});

mail.outgoingMessages.push(reply);

// Add original sender as To recipient
var toRecip = mail.Recipient({ address: originalSender });
reply.toRecipients.push(toRecip);

// For reply-all, add original To and CC (excluding self)
if (replyAll) {
  try {
    var origTo = msg.toRecipients();
    for (var i = 0; i < origTo.length; i++) {
      var addr = origTo[i].address();
      if (addr !== originalSender) {
        reply.toRecipients.push(mail.Recipient({ address: addr }));
      }
    }
  } catch (e) {}

  try {
    var origCc = msg.ccRecipients();
    for (var i = 0; i < origCc.length; i++) {
      reply.ccRecipients.push(
        mail.CcRecipient({ address: origCc[i].address() })
      );
    }
  } catch (e) {}
}

if (shouldSend) {
  reply.send();
  JSON.stringify({
    success: true,
    message: "Reply sent successfully.",
  });
} else {
  JSON.stringify({
    success: true,
    message: "Reply draft created and opened for review.",
  });
}
