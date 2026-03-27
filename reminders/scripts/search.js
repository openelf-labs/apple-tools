ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var app = Application("Reminders");
app.includeStandardAdditions = true;

var query = (params.query || "").toLowerCase();
var limit = params.limit || 20;
var matchAll = !query || query.length === 0;

var results = [];
var lists = app.lists();

for (var li = 0; li < lists.length; li++) {
  var list = lists[li];
  var rems = list.reminders();

  for (var ri = 0; ri < rems.length; ri++) {
    if (results.length >= limit) break;

    var rem = rems[ri];
    var title = rem.name() || "";
    var notes = rem.body() || "";

    if (
      !matchAll &&
      title.toLowerCase().indexOf(query) === -1 &&
      notes.toLowerCase().indexOf(query) === -1
    ) {
      continue;
    }

    var dueDate = rem.dueDate();
    var creationDate = rem.creationDate();

    results.push({
      id: rem.id(),
      title: title,
      notes: notes,
      dueDate: dueDate ? dueDate.toISOString() : null,
      completed: rem.completed(),
      priority: rem.priority(),
      list: list.name(),
      creationDate: creationDate ? creationDate.toISOString() : null,
    });
  }

  if (results.length >= limit) break;
}

JSON.stringify(results);
