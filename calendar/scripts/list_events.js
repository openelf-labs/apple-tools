ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var cal = Application("Calendar");
cal.includeStandardAdditions = true;

var fromDate = params.from ? new Date(params.from) : new Date();
var toDate;
if (params.to) {
  toDate = new Date(params.to);
} else {
  toDate = new Date(fromDate.getTime());
  toDate.setDate(toDate.getDate() + 7);
}

var limit = params.limit || 20;
var calendarFilter = params.calendar || "";

var results = [];
var calendars = cal.calendars();

for (var ci = 0; ci < calendars.length; ci++) {
  var c = calendars[ci];
  if (calendarFilter && c.name() !== calendarFilter) {
    continue;
  }

  var events = c.events.whose({
    _and: [
      { startDate: { _greaterThan: fromDate } },
      { startDate: { _lessThan: toDate } },
    ],
  })();

  for (var ei = 0; ei < events.length; ei++) {
    var ev = events[ei];
    try {
      var loc = "";
      try { loc = ev.location() || ""; } catch (e) {}
      var n = "";
      try { n = ev.description() || ""; } catch (e) {}
      var u = "";
      try { u = ev.url() || ""; } catch (e) {}

      results.push({
        id: ev.uid(),
        title: ev.summary(),
        startDate: ev.startDate().toISOString(),
        endDate: ev.endDate().toISOString(),
        location: loc,
        notes: n,
        calendar: c.name(),
        allDay: ev.alldayEvent(),
        url: u,
      });
    } catch (e) {
      // skip malformed events
    }
  }
}

// Sort by start date ascending
results.sort(function (a, b) {
  return new Date(a.startDate) - new Date(b.startDate);
});

// Apply limit
results = results.slice(0, limit);

JSON.stringify(results);
