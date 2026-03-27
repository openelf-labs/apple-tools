ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var mail = Application("Mail");
mail.includeStandardAdditions = true;

var targetId = params.messageId;

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

var changes = [];

if (params.isRead !== undefined && params.isRead !== null) {
  msg.readStatus = params.isRead;
  changes.push(params.isRead ? "marked as read" : "marked as unread");
}

if (params.isFlagged !== undefined && params.isFlagged !== null) {
  msg.flaggedStatus = params.isFlagged;
  changes.push(params.isFlagged ? "flagged" : "unflagged");
}

if (params.isJunk !== undefined && params.isJunk !== null) {
  msg.junkMailStatus = params.isJunk;
  changes.push(params.isJunk ? "marked as junk" : "marked as not junk");
}

var message;
if (changes.length === 0) {
  message = "No changes specified.";
} else {
  message = "Message " + changes.join(", ") + ".";
}

JSON.stringify({
  success: changes.length > 0,
  message: message,
});
