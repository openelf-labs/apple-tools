ObjC.import("Foundation");

var mail = Application("Mail");
mail.includeStandardAdditions = true;

var results = [];
var accounts = mail.accounts();

function collectMailboxes(mailboxes, acctName, parentPath) {
  for (var mi = 0; mi < mailboxes.length; mi++) {
    try {
      var mb = mailboxes[mi];
      var mbName = mb.name();
      var path = parentPath ? parentPath + "/" + mbName : acctName + "/" + mbName;
      var unread = 0;
      var total = 0;
      try { unread = mb.unreadCount(); } catch (e) {}
      try { total = mb.count(); } catch (e) {}

      results.push({
        path: path,
        account: acctName,
        name: parentPath ? path.substring(acctName.length + 1) : mbName,
        unreadCount: unread,
        messageCount: total,
      });

      // Recurse into child mailboxes
      try {
        var children = mb.mailboxes();
        if (children.length > 0) {
          collectMailboxes(children, acctName, path);
        }
      } catch (e) {}
    } catch (e) {
      // skip inaccessible mailboxes
    }
  }
}

for (var ai = 0; ai < accounts.length; ai++) {
  var acct = accounts[ai];
  var acctName = acct.name();
  var mailboxes = acct.mailboxes();
  collectMailboxes(mailboxes, acctName, "");
}

JSON.stringify(results);
