ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var app = Application("Reminders");
app.includeStandardAdditions = true;

var title = params.title || "";
if (!title) {
  throw new Error("title parameter is required");
}

var listName = params.list || "";
var notes = params.notes || "";
var dueDateStr = params.dueDate || "";
var priority = params.priority || 0;

// Find or default to the first list.
var targetList;
if (listName) {
  try {
    targetList = app.lists.byName(listName);
    // Force access to verify the list exists.
    targetList.name();
  } catch (e) {
    throw new Error("Reminder list not found: " + listName);
  }
} else {
  targetList = app.defaultList();
}

// Build reminder properties.
var props = { name: title };

if (notes) {
  props.body = notes;
}

if (dueDateStr) {
  props.dueDate = new Date(dueDateStr);
}

if (priority >= 1 && priority <= 9) {
  props.priority = priority;
}

var rem = app.Reminder(props);
targetList.reminders.push(rem);

var dueDate = rem.dueDate();

JSON.stringify({
  success: true,
  id: rem.id(),
  title: rem.name(),
  dueDate: dueDate ? dueDate.toISOString() : null,
  list: targetList.name(),
});
