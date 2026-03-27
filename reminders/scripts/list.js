ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var app = Application("Reminders");
app.includeStandardAdditions = true;

var listName = params.list || "";
var status = params.status || "incomplete";
var limit = params.limit || 50;
var dueAfter = params.dueAfter || "";
var dueBefore = params.dueBefore || "";

var lists;
if (listName) {
  try {
    lists = [app.lists.byName(listName)];
    // Force access to verify the list exists.
    lists[0].name();
  } catch (e) {
    throw new Error("Reminder list not found: " + listName);
  }
} else {
  lists = app.lists();
}

var dueAfterDate = dueAfter ? new Date(dueAfter) : null;
var dueBeforeDate = dueBefore ? new Date(dueBefore) : null;

var results = [];

for (var li = 0; li < lists.length; li++) {
  var list = lists[li];
  var rems = list.reminders();

  for (var ri = 0; ri < rems.length; ri++) {
    if (results.length >= limit) break;

    var rem = rems[ri];
    var completed = rem.completed();

    if (status === "incomplete" && completed) continue;
    if (status === "completed" && !completed) continue;

    var dueDate = rem.dueDate();
    var dueDateISO = null;
    if (dueDate) {
      dueDateISO = dueDate.toISOString();

      if (dueAfterDate && dueDate < dueAfterDate) continue;
      if (dueBeforeDate && dueDate > dueBeforeDate) continue;
    } else {
      // If filtering by due date range, skip reminders without due dates.
      if (dueAfterDate || dueBeforeDate) continue;
    }

    var creationDate = rem.creationDate();

    results.push({
      id: rem.id(),
      title: rem.name(),
      notes: rem.body() || "",
      dueDate: dueDateISO,
      completed: completed,
      priority: rem.priority(),
      list: list.name(),
      creationDate: creationDate ? creationDate.toISOString() : null,
    });
  }

  if (results.length >= limit) break;
}

JSON.stringify(results);
