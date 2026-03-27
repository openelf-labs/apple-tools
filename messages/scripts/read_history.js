ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var phone = params.phoneNumber;
var limit = params.limit || 10;

// Normalize phone number: strip spaces, dashes, parens for matching
var normalizedPhone = phone.replace(/[\s\-\(\)]/g, "");

// Build a LIKE pattern that matches various phone formats
// E.g., +15551234567 should match 5551234567, +1 555 123 4567, etc.
var likePhone = "%" + normalizedPhone.replace(/^\+?1/, "") + "%";

var app = Application.currentApplication();
app.includeStandardAdditions = true;

var dbPath = app.doShellScript("echo ~/Library/Messages/chat.db").replace(/\n$/, "");

var sql = "SELECT " +
  "m.text, " +
  "datetime(m.date/1000000000 + 978307200, 'unixepoch', 'localtime') as msg_date, " +
  "m.is_from_me, " +
  "COALESCE(h.id, 'me') as sender " +
  "FROM message m " +
  "LEFT JOIN handle h ON m.handle_id = h.ROWID " +
  "LEFT JOIN chat_message_join cmj ON m.ROWID = cmj.message_id " +
  "LEFT JOIN chat c ON cmj.chat_id = c.ROWID " +
  "WHERE (h.id LIKE '" + likePhone.replace(/'/g, "''") + "' " +
  "OR c.chat_identifier LIKE '" + likePhone.replace(/'/g, "''") + "') " +
  "AND m.text IS NOT NULL " +
  "ORDER BY m.date DESC " +
  "LIMIT " + limit;

var output = app.doShellScript("sqlite3 -json '" + dbPath.replace(/'/g, "'\\''") + "' " + JSON.stringify(sql));

var rows = [];
if (output && output.trim().length > 0) {
  try {
    rows = JSON.parse(output);
  } catch (e) {
    rows = [];
  }
}

var messages = [];
for (var i = 0; i < rows.length; i++) {
  var row = rows[i];
  messages.push({
    content: row.text || "",
    date: row.msg_date || "",
    isFromMe: row.is_from_me === 1,
    sender: row.is_from_me === 1 ? "me" : (row.sender || phone),
  });
}

JSON.stringify(messages);
