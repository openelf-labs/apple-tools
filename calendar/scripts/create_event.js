ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var cal = Application("Calendar");
cal.includeStandardAdditions = true;

// Find the target calendar
var calendarName = params.calendar || "";
var targetCalendar = null;

if (calendarName) {
  var calendars = cal.calendars();
  for (var i = 0; i < calendars.length; i++) {
    if (calendars[i].name() === calendarName) {
      targetCalendar = calendars[i];
      break;
    }
  }
  if (!targetCalendar) {
    throw new Error("Calendar not found: " + calendarName);
  }
} else {
  // Use the default calendar
  targetCalendar = cal.defaultCalendar();
}

var startDate = new Date(params.start);
var endDate;

if (params.end) {
  endDate = new Date(params.end);
} else if (params.allDay) {
  // All-day events default to same day
  endDate = new Date(startDate.getTime());
  endDate.setDate(endDate.getDate() + 1);
} else {
  // Default to 1 hour duration
  endDate = new Date(startDate.getTime() + 60 * 60 * 1000);
}

var props = {
  summary: params.title,
  startDate: startDate,
  endDate: endDate,
  alldayEvent: params.allDay || false,
};

if (params.location) {
  props.location = params.location;
}
if (params.notes) {
  props.description = params.notes;
}
if (params.url) {
  props.url = params.url;
}

var ev = cal.Event(props);
targetCalendar.events.push(ev);

// Add alert if requested
if (params.alertMinutes !== undefined && params.alertMinutes !== null) {
  var alarm = cal.DisplayAlarm({
    triggerInterval: -(params.alertMinutes * 60),
  });
  ev.displayAlarms.push(alarm);
}

JSON.stringify({
  success: true,
  eventId: ev.uid(),
  title: ev.summary(),
  startDate: ev.startDate().toISOString(),
  endDate: ev.endDate().toISOString(),
});
