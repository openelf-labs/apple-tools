ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var mail = Application("Mail");
mail.includeStandardAdditions = true;

var query = (params.query || "").toLowerCase();
var matchAll = !query || query.length === 0;
var limit = params.limit || 10;
var sinceDate = params.since ? new Date(params.since) : null;

// Optional: restrict to a specific mailbox path
var mailboxPath = params.mailbox || "";
var targetAccount = "";
var targetMbParts = [];
if (mailboxPath) {
  var parts = mailboxPath.split("/");
  targetAccount = parts[0];
  targetMbParts = parts.slice(1);
}

// Find a specific nested mailbox by path parts within an account
function findMailbox(accountMailboxes, mbParts) {
  var mailboxes = accountMailboxes;
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

// Search messages in a single mailbox
function searchMailbox(mb, results) {
  try {
    var messages = mb.messages();
    for (var i = 0; i < messages.length && results.length < limit; i++) {
      try {
        var msg = messages[i];
        var dateReceived = msg.dateReceived();

        if (sinceDate && dateReceived < sinceDate) {
          continue;
        }

        var subject = msg.subject() || "";
        var sender = msg.sender() || "";
        var content = "";
        try { content = msg.content() || ""; } catch (e) {}

        if (
          !matchAll &&
          subject.toLowerCase().indexOf(query) === -1 &&
          sender.toLowerCase().indexOf(query) === -1 &&
          content.toLowerCase().indexOf(query) === -1
        ) {
          continue;
        }

        var hasAttachments = false;
        try { hasAttachments = msg.mailAttachments().length > 0; } catch (e) {}

        results.push({
          messageId: msg.messageId(),
          subject: subject || "(no subject)",
          sender: sender,
          dateReceived: dateReceived.toISOString(),
          isRead: msg.readStatus(),
          isFlagged: msg.flaggedStatus(),
          hasAttachments: hasAttachments,
        });
      } catch (e) {}
    }
  } catch (e) {}
}

// Recursively search all mailboxes
function searchAllMailboxes(mailboxes, results) {
  for (var mi = 0; mi < mailboxes.length && results.length < limit; mi++) {
    searchMailbox(mailboxes[mi], results);
    try {
      var children = mailboxes[mi].mailboxes();
      if (children.length > 0) {
        searchAllMailboxes(children, results);
      }
    } catch (e) {}
  }
}

var results = [];
var accounts = mail.accounts();

for (var ai = 0; ai < accounts.length && results.length < limit; ai++) {
  var acct = accounts[ai];
  if (targetAccount && acct.name() !== targetAccount) {
    continue;
  }

  if (targetMbParts.length > 0) {
    var mb = findMailbox(acct.mailboxes(), targetMbParts);
    if (mb) {
      searchMailbox(mb, results);
    }
  } else {
    searchAllMailboxes(acct.mailboxes(), results);
  }
}

var hasMore = false;

JSON.stringify({
  messages: results,
  total: results.length,
  hasMore: hasMore,
});
