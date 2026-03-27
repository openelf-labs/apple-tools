ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var cal = Application("Calendar");
cal.includeStandardAdditions = true;

var query = (params.query || "").toLowerCase();
var matchAll = !query || query.length === 0;
var limit = params.limit || 10;

// Default search range: past 30 days to next 90 days
var now = new Date();
var fromDate = params.from
  ? new Date(params.from)
  : new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
var toDate = params.to
  ? new Date(params.to)
  : new Date(now.getTime() + 90 * 24 * 60 * 60 * 1000);

var results = [];
var calendars = cal.calendars();

for (var ci = 0; ci < calendars.length; ci++) {
  var c = calendars[ci];

  var events = c.events.whose({
    _and: [
      { startDate: { _greaterThan: fromDate } },
      { startDate: { _lessThan: toDate } },
    ],
  })();

  for (var ei = 0; ei < events.length; ei++) {
    var ev = events[ei];
    try {
      var title = ev.summary() || "";
      if (!matchAll && title.toLowerCase().indexOf(query) === -1) {
        continue;
      }

      var loc = "";
      try { loc = ev.location() || ""; } catch (e) {}
      var n = "";
      try { n = ev.description() || ""; } catch (e) {}
      var u = "";
      try { u = ev.url() || ""; } catch (e) {}

      results.push({
        id: ev.uid(),
        title: title,
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
