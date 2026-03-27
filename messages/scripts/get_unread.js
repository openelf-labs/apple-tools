ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var limit = params.limit || 10;

var app = Application.currentApplication();
app.includeStandardAdditions = true;

var dbPath = app.doShellScript("echo ~/Library/Messages/chat.db").replace(/\n$/, "");

var sql = "SELECT " +
  "m.text, " +
  "datetime(m.date/1000000000 + 978307200, 'unixepoch', 'localtime') as msg_date, " +
  "COALESCE(h.id, 'unknown') as sender, " +
  "c.chat_identifier as chat_id " +
  "FROM message m " +
  "LEFT JOIN handle h ON m.handle_id = h.ROWID " +
  "LEFT JOIN chat_message_join cmj ON m.ROWID = cmj.message_id " +
  "LEFT JOIN chat c ON cmj.chat_id = c.ROWID " +
  "WHERE m.is_read = 0 " +
  "AND m.is_from_me = 0 " +
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
    sender: row.sender || "unknown",
    chatId: row.chat_id || "",
  });
}

JSON.stringify(messages);
