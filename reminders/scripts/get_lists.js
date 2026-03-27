ObjC.import("Foundation");

var app = Application("Reminders");
app.includeStandardAdditions = true;

var lists = app.lists();
var results = [];

for (var i = 0; i < lists.length; i++) {
  var list = lists[i];
  results.push({
    id: list.id(),
    name: list.name(),
    color: list.color() || null,
  });
}

JSON.stringify(results);
