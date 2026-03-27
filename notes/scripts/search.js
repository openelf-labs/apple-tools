ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var app = Application("Notes");
app.includeStandardAdditions = true;

var query = (params.query || "").toLowerCase();
var limit = params.limit || 50;

var results = [];
var folders = app.folders();

for (var fi = 0; fi < folders.length; fi++) {
  var folder = folders[fi];
  var notes = folder.notes();

  for (var ni = 0; ni < notes.length; ni++) {
    if (results.length >= limit) break;

    var note = notes[ni];
    try {
      var name = note.name() || "";
      var body = "";
      try { body = note.plaintext() || ""; } catch (e) {}

      if (
        name.toLowerCase().indexOf(query) !== -1 ||
        body.toLowerCase().indexOf(query) !== -1
      ) {
        var snippet = body.substring(0, 200);

        results.push({
          name: name,
          folder: folder.name(),
          snippet: snippet,
          creationDate: note.creationDate().toISOString(),
          modificationDate: note.modificationDate().toISOString(),
        });
      }
    } catch (e) {
      // skip malformed notes
    }
  }

  if (results.length >= limit) break;
}

// Sort by modification date descending
results.sort(function (a, b) {
  return new Date(b.modificationDate) - new Date(a.modificationDate);
});

results = results.slice(0, limit);

JSON.stringify(results);
