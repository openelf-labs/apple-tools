ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var mail = Application("Mail");
mail.includeStandardAdditions = true;

var mailboxPath = params.mailbox;
var limit = params.limit || 20;
var offset = params.offset || 0;
var unreadOnly = params.unreadOnly || false;
var sinceDate = params.since ? new Date(params.since) : null;
var fromFilter = params.from ? params.from.toLowerCase() : "";

// Find the mailbox by path (account/mailbox or account/parent/child/...)
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

var targetMailbox = findMailboxByPath(mailboxPath);

if (!targetMailbox) {
  throw new Error("Mailbox not found: " + mailboxPath);
}

var messages = targetMailbox.messages();
var total = messages.length;
var results = [];
var count = 0;
var skipped = 0;

// Messages are typically in reverse chronological order
for (var i = 0; i < messages.length && count < limit; i++) {
  try {
    var msg = messages[i];
    var dateReceived = msg.dateReceived();

    // Apply since filter
    if (sinceDate && dateReceived < sinceDate) {
      continue;
    }

    // Apply unread filter
    var readStatus = msg.readStatus();
    if (unreadOnly && readStatus) {
      continue;
    }

    // Apply sender filter
    var sender = msg.sender();
    if (fromFilter && sender.toLowerCase().indexOf(fromFilter) === -1) {
      continue;
    }

    // Apply offset
    if (skipped < offset) {
      skipped++;
      continue;
    }

    var hasAttachments = false;
    try { hasAttachments = msg.mailAttachments().length > 0; } catch (e) {}

    results.push({
      messageId: msg.messageId(),
      subject: msg.subject() || "(no subject)",
      sender: sender,
      dateReceived: dateReceived.toISOString(),
      isRead: readStatus,
      isFlagged: msg.flaggedStatus(),
      hasAttachments: hasAttachments,
    });
    count++;
  } catch (e) {
    // skip malformed messages
  }
}

var hasMore = (offset + count) < total;

JSON.stringify({
  messages: results,
  total: total,
  hasMore: hasMore,
});
