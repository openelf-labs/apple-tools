ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var cal = Application("Calendar");
cal.includeStandardAdditions = true;

var eventId = params.eventId;
var targetEvent = null;
var calendars = cal.calendars();

for (var ci = 0; ci < calendars.length; ci++) {
  var events = calendars[ci].events.whose({ uid: eventId })();
  if (events.length > 0) {
    targetEvent = events[0];
    break;
  }
}

if (!targetEvent) {
  throw new Error("Event not found: " + eventId);
}

// Navigate Calendar to the event's date
var eventDate = targetEvent.startDate();
cal.activate();
cal.viewCalendar({ at: eventDate });

JSON.stringify({
  success: true,
  message: "Opened Calendar at " + eventDate.toISOString(),
});
