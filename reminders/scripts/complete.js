ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var app = Application("Reminders");
app.includeStandardAdditions = true;

var id = params.id || "";
if (!id) {
  throw new Error("id parameter is required");
}

// Search across all lists for a reminder matching by id or name.
var found = null;
var lists = app.lists();

for (var li = 0; li < lists.length; li++) {
  var rems = lists[li].reminders();
  for (var ri = 0; ri < rems.length; ri++) {
    var rem = rems[ri];
    if (rem.id() === id || rem.name() === id) {
      found = rem;
      break;
    }
  }
  if (found) break;
}

if (!found) {
  throw new Error("Reminder not found: " + id);
}

if (found.completed()) {
  JSON.stringify({
    success: true,
    title: found.name(),
    message: "Reminder was already completed",
  });
} else {
  found.completed = true;

  JSON.stringify({
    success: true,
    title: found.name(),
    message: "Reminder marked as completed",
  });
}
