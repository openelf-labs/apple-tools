ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var mail = Application("Mail");
mail.includeStandardAdditions = true;

var targetId = params.messageId;
var destPath = params.destinationMailbox;

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

// Find a mailbox by nested path (account/parent/child/...)
function findMailboxByPath(path) {
  var parts = path.split("/");
  var accountName = parts[0];
  var mbParts = parts.slice(1);

  var accounts = mail.accounts();
  for (var ai = 0; ai < accounts.length; ai++) {
    if (accounts[ai].name() !== accountName) continue;

    var mailboxes = accounts[ai].mailboxes();
    var current = null;

    for (var depth = 0; depth < mbParts.length; depth++) {
      var found = false;
      for (var mi = 0; mi < mailboxes.length; mi++) {
        if (mailboxes[mi].name() === mbParts[depth]) {
          current = mailboxes[mi];
          if (depth < mbParts.length - 1) {
            mailboxes = current.mailboxes();
          }
          found = true;
          break;
        }
      }
      if (!found) return null;
    }
    return current;
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

var destMailbox = findMailboxByPath(destPath);

if (!destMailbox) {
  throw new Error("Destination mailbox not found: " + destPath);
}

// Move the message
msg.mailbox = destMailbox;

JSON.stringify({
  success: true,
  message: "Message moved to " + destPath + ".",
});
